package downloader

import (
	"fmt"
	"log"
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
	}
)

func NewTorrent(peerID [20]byte, torrentInfo model.TorrentInfo, timeout time.Duration) (*Torrent, error) {
	return &Torrent{
		peerID:      peerID,
		torrentInfo: torrentInfo,
		timeout:     timeout,
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

func (t *Torrent) Download() {
	for peer := range t.peers.peersChan {
		go func(p *Peer) {
			if err := t.download(p); err != nil {
				log.Printf("faield to download from peer, peerIP: %s, err: %s", p.ip, err)
			}
		}(peer)
	}
}

func (t *Torrent) download(peer *Peer) error {
	defer func() {
		if err := t.peers.removePeerIP(peer.ip); err != nil {
			log.Printf("failed to remove peerIP: %s", err.Error())
		}
	}()

	return nil
}
