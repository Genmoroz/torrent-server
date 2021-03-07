package bencode

import (
	"fmt"
	"io"

	"github.com/genvmoroz/simple-torrent-client/model"
	"github.com/jackpal/bencode-go"
)

func ParseTorrentInfo(r io.Reader) (model.TorrentInfo, error) {
	b := bitTorrent{}
	if err := bencode.Unmarshal(r, &b); err != nil {
		return model.TorrentInfo{}, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return toDomainBitTorrent(b)
}

func ParseTrackerInfo(r io.Reader) (model.TrackerInfo, error) {
	tr := trackerResponse{}
	if err := bencode.Unmarshal(r, &tr); err != nil {
		return model.TrackerInfo{}, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return toDomainTrackerInfoWithoutPeersInfo(tr)
}
