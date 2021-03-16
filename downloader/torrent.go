package downloader

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/genvmoroz/simple-torrent-client/client"
	"github.com/genvmoroz/simple-torrent-client/model"
)

const tcp = "tcp"

type (
	Torrent struct {
		peerID      [20]byte
		torrentInfo model.TorrentInfo
		timeout     time.Duration
		peers       Peers
		workPool    chan *workPiece
	}

	workPiece struct {
		index  int
		hash   [20]byte
		length int64
	}

	workResult struct {
		index int
		buf   []byte
	}
)

func calculateWorkPieceLength(index, pieceLength, fileLength int64) int64 {
	begin := index * pieceLength
	end := begin + pieceLength
	if end > fileLength {
		end = fileLength
	}

	return end - begin
}

func NewTorrent(peerID [20]byte, torrentInfo model.TorrentInfo, timeout time.Duration) (*Torrent, error) {
	workPool := make(chan *workPiece, len(torrentInfo.PieceHashes))
	for index, hash := range torrentInfo.PieceHashes {
		workPool <- &workPiece{
			index,
			hash,
			calculateWorkPieceLength(int64(index), torrentInfo.PieceLength, torrentInfo.Length),
		}
	}

	return &Torrent{
		peerID:      peerID,
		torrentInfo: torrentInfo,
		timeout:     timeout,
		workPool:    workPool,
		peers: Peers{
			peerIPs:   make([]string, 0),
			mux:       sync.Mutex{},
			peersChan: make(chan *Peer, 1024),
		},
	}, nil
}

func (t *Torrent) ConnectToPeers() error {
	trackerInfo, err := client.GetTrackerInfo(t.torrentInfo, t.peerID)
	if err != nil {
		return fmt.Errorf("failed to get TrackerInfo: %w", err)
	}

	for _, peerInfo := range trackerInfo.Peers {
		go func(pi model.PeerInfo) {
			if err = t.connectToPeer(pi); err != nil {
				log.Printf("failed to connect to peer, peerIP: %s, err: %s", pi.IP.String(), err)
			}
		}(peerInfo)
	}

	return nil
}

func (t *Torrent) Download(ctx context.Context) error {
	file, err := os.Create(t.torrentInfo.Name)
	if err != nil {
		return fmt.Errorf("failed to download a file, fileName; %s, err: %w", t.torrentInfo.Name, err)
	}

	resultChan := make(chan *workResult, len(t.torrentInfo.PieceHashes))

	for {
		select {
		case <-ctx.Done():
			return nil
		case peer := <-t.peers.peersChan:
			go func(p *Peer) {
				if err = t.download(ctx, p, resultChan); err != nil {
					log.Printf("faield to download from peer, peerIP: %s, err: %s", p.ip, err.Error())
				}
			}(peer)
		case result := <-resultChan:
			go func(f *os.File, r *workResult) {
				if err = t.writeFile(f, r); err != nil {
					log.Printf("faield to write file: %s", err.Error())
				}
			}(file, result)
		}
	}
}

func (t *Torrent) connectToPeer(peerInfo model.PeerInfo) error {
	if t.peers.existPeerIP(peerInfo.IP.String()) {
		log.Println("the port with such portIP is already presented, return")
		return nil
	}

	peer, err := ConnectToPeer(tcp, peerInfo.IP.String(), peerInfo.Port, t.torrentInfo.InfoHash, t.peerID)
	if err != nil {
		log.Printf("failed to connect to Peer: %s", err.Error())
	} else {
		if err = t.peers.addPeer(peerInfo.IP.String(), peer); err != nil {
			log.Printf("failed to add peer for torrent, name: %s, err: %s", t.torrentInfo.Name, err.Error())
		}
	}

	return nil
}

func (t *Torrent) writeFile(f *os.File, r *workResult) error {
	if f == nil {
		return errors.New("file cannot be nil")
	}
	if r == nil {
		return errors.New("workResult cannot be nil")
	}

	number, err := f.WriteAt(r.buf, t.torrentInfo.PieceLength*int64(r.index))
	if err != nil {
		return fmt.Errorf("failed to write piece into file: %w", err)
	}

	log.Printf("written %d bytes", number)

	return nil
}

func (t *Torrent) download(ctx context.Context, peer *Peer, resultChan chan *workResult) error {
	if peer == nil {
		return errors.New("peer cannot be nil")
	}
	defer func() {
		if err := t.peers.removePeerIP(peer.ip); err != nil {
			log.Printf("failed to remove peerIP: %s", err.Error())
		}
	}()
	var maxBacklog int64 = 5
	for {
		select {
		case <-ctx.Done():
			return nil
		case work := <-t.workPool:
			downloadBuff := make([]byte, work.length)

			var downloaded, requested, backlog int64 = 0, 0, 0

			for int64(downloaded) < work.length {
				if !peer.choked {
					for backlog < maxBacklog && (requested) < work.length {
						blockSize := maxBacklog
						// Last block might be shorter than the typical block
						if work.length-int64(requested) < int64(blockSize) {
							blockSize = work.length - int64(requested)
						}

						err := peer.SendRequest(int64(work.index), requested, blockSize)
						if err != nil {
							return err
						}
						backlog++
						requested += blockSize
					}
				}

				err := state.readMessage()
				if err != nil {
					return nil, err
				}
			}
			// todo: do download
			resultChan <- &workResult{
				index: work.index,
				buf:   downloadBuff,
			}
		}
	}

	return nil
}
