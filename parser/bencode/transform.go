package bencode

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/genvmoroz/simple-torrent-client/model"
	"github.com/jackpal/bencode-go"
)

const (
	hashLen  = 20 // Length of SHA-1 hash
	peerSize = 6  // 4 for IP, 2 for port
)

func toDomainBitTorrent(torrent bitTorrent) (model.TorrentInfo, error) {
	pieceHashes, err := splitPieceHashes(torrent.Info.Pieces)
	if err != nil {
		return model.TorrentInfo{}, fmt.Errorf("failed to split piece hashes: %w", err)
	}

	infoHash, err := infoHash(torrent.Info)

	return model.TorrentInfo{
		Announce:     torrent.Announce,
		AnnounceList: torrent.AnnounceList,
		Comment:      torrent.Comment,
		CreatedBy:    torrent.CreatedBy,
		CreationDate: time.Unix(torrent.CreationDate, 0),
		Encoding:     torrent.Encoding,
		InfoHash:     infoHash,
		PieceHashes:  pieceHashes,
		PieceLength:  torrent.Info.PieceLength,
		Length:       torrent.Info.Length,
		Name:         torrent.Info.Name,
	}, nil
}

func toDomainTrackerInfoWithoutPeersInfo(tr trackerResponse) (model.TrackerInfo, error) {
	peers, err := parsePeers([]byte(tr.Peers))
	if err != nil {
		return model.TrackerInfo{}, fmt.Errorf("failed to parse Peers: %w", err)
	}

	return model.TrackerInfo{
		Interval: tr.Interval,
		Peers:    peers,
	}, nil
}

func parsePeers(rawPeers []byte) ([]model.PeerInfo, error) {
	numPeers := len(rawPeers) / peerSize
	if len(rawPeers)%peerSize != 0 {
		return nil, fmt.Errorf("received malformed peers")
	}

	peers := make([]model.PeerInfo, numPeers)
	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		peers[i].IP = rawPeers[offset : offset+4]
		peers[i].Port = binary.BigEndian.Uint16(rawPeers[offset+4 : offset+6])
	}

	return peers, nil
}

func splitPieceHashes(s string) ([][20]byte, error) {
	buf := []byte(s)
	if len(buf)%hashLen != 0 {
		return nil, fmt.Errorf("received malformed pieces of length %d", len(buf))
	}

	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)
	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}

	return hashes, nil
}

func infoHash(i info) ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, i)
	if err != nil {
		return [20]byte{}, fmt.Errorf("failed to marshal: %w", err)
	}

	return sha1.Sum(buf.Bytes()), nil
}
