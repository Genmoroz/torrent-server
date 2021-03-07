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
		PieceLength  int64
		Length       int64
		Name         string
	}

	TrackerInfo struct {
		Interval int64
		Peers    []PeerInfo
	}

	PeerInfo struct {
		IP   net.IP
		Port uint16
	}
)
