package downloader

import (
	"fmt"
	"log"
	"time"

	"github.com/genvmoroz/simple-torrent-client/model"
)

type TorrentDownloader struct {
	peerID   [20]byte // shouldn't be changed
	torrents []*Torrent
}

func NewTorrentDownloader(peerID [20]byte, torrentInfo []model.TorrentInfo, timeout time.Duration) (*TorrentDownloader, error) {
	torrents := make([]*Torrent, len(torrentInfo))

	for index, ti := range torrentInfo {
		torrent, err := NewTorrent(peerID, ti, timeout)
		if err != nil {
			return nil, fmt.Errorf("failed to create a new Torrent: %w", err)
		}
		torrents[index] = torrent
	}

	return &TorrentDownloader{
		peerID:   peerID,
		torrents: torrents,
	}, nil
}

func (d *TorrentDownloader) Download() error {
	ticker := time.NewTicker(time.Minute)

	for {
		for _, torrent := range d.torrents {
			if torrent == nil {
				continue
			}
			go func(t *Torrent) {
				if err := t.ConnectToPeers(); err != nil {
					log.Printf("failed to connect to peers and download, torrent name: %s, err: %s", t.torrentInfo.Name, err.Error())
				}
			}(torrent)
		}

		select {

		case <-ticker.C:
		}
	}
}
