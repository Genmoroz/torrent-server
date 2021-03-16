package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/genvmoroz/simple-torrent-client/downloader"
	"github.com/genvmoroz/simple-torrent-client/loader"
	"github.com/genvmoroz/simple-torrent-client/model"
	"github.com/genvmoroz/simple-torrent-client/parser/bencode"
)

func main() {
	content, err := loader.ReadFile("./test.torrent")
	if err != nil {
		log.Fatal(err)
	}

	peerID := [20]byte{}
	_, err = rand.Read(peerID[:])
	if err != nil {
		log.Fatalln(err)
	}

	torrentInfo, err := bencode.ParseTorrentInfo(content)

	torrentDownloader, err := downloader.NewTorrentDownloader(peerID, []model.TorrentInfo{torrentInfo}, 10*time.Second)
	if err != nil {
		log.Fatalln(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Hour)

	if err = torrentDownloader.Download(ctx); err != nil {
		log.Fatalln(err)
	}
}
