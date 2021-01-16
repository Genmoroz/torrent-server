package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"torrent-server/model"
)

func Get(url string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	fmt.Println(res)

	return nil
}

func PrepareTrackerURL(bitTorrent model.BitTorrent, peerID [20]byte, port uint16) (*url.URL, error) {
	base, err := url.Parse(bitTorrent.Announce)
	if err != nil {
		return nil, err
	}
	params := url.Values{
		"info_hash":  []string{string(bitTorrent.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(bitTorrent.Length)},
	}
	base.RawQuery = params.Encode()
	return base, nil
}
