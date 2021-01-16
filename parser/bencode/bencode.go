package bencode

import (
	"fmt"
	"io"

	"github.com/jackpal/bencode-go"
	"torrent-server/model"
	"torrent-server/parser"
)

type benCodeParser struct{}

func NewParser() parser.Parser {
	return &benCodeParser{}
}

func (*benCodeParser) Parse(r io.Reader) (model.BitTorrent, error) {
	b := bitTorrent{}
	if err := bencode.Unmarshal(r, &b); err != nil {
		return model.BitTorrent{}, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return toDomainModel(b), nil
}
