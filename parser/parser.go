package parser

import (
	"io"

	"torrent-server/model"
)

type Parser interface {
	Parse(r io.Reader) (model.BitTorrent, error)
}
