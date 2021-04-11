package model

import (
	"net"
	"time"
)

type (
	TorrentInfo struct {
		Announce     string
		AnnounceList [][]string
		Comment      string
		CreatedBy    string
		CreationDate time.Time
		Encoding     string
		InfoHash     [20]byte
		PieceHashes  [][20]byte
		PieceLength  uint32
		Length       uint32
		Name         string
	}

	TrackerInfo struct {
		Interval uint32
		Peers    []PeerInfo
	}

	PeerInfo struct {
		IP   net.IP
		Port uint16
	}
)
