package model

import "time"

type (
	BitTorrent struct {
		Announce     string
		AnnounceList string
		Comment      string
		CreatedBy    string
		CreationDate time.Time
		Encoding     string
		Info         Info
	}

	Info struct {
		Pieces      string
		PieceLength int
		Length      int
		Name        string
	}
)
