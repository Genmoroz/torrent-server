package model

import "time"

type (
	BitTorrent struct {
		Announce     string
		AnnounceList [][]string
		Comment      string
		CreatedBy    string
		CreationDate time.Time
		Encoding     string
		InfoHash     [20]byte
		PieceHashes  [][20]byte
		PieceLength  int
		Length       int
		Name         string
	}
)
