package client

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/genvmoroz/simple-torrent-client/model"
	"github.com/genvmoroz/simple-torrent-client/parser/bencode"
)

const port = 6881

func GetTrackerInfo(torrentInfo model.TorrentInfo, peerID [20]byte) (model.TrackerInfo, error) {
	announces := make([]string, 0)
	for _, announceArray := range torrentInfo.AnnounceList {
		for _, announce := range announceArray {
			announces = append(announces, announce)
		}
	}
	if len(announces) == 0 {
		return model.TrackerInfo{}, errors.New("announces cannot be empty")
	}

	peers := make([]model.PeerInfo, 0)

	trackerInfo, err := getTrackerInfo(torrentInfo.InfoHash, peerID, announces[0], torrentInfo.Length, port)
	if err != nil {
		return model.TrackerInfo{}, fmt.Errorf("with announce: %s, err: %w", announces[0], err)
	}

	for _, peer := range trackerInfo.Peers {
		peers = appendWithoutDuplicates(peers, peer)
	}

	for i := 1; i < len(announces); i++ {
		var ti model.TrackerInfo
		ti, err = getTrackerInfo(torrentInfo.InfoHash, peerID, announces[0], torrentInfo.Length, port)
		if err != nil {
			return model.TrackerInfo{}, fmt.Errorf("with announce: %s, err: %w", announces[0], err)
		}
		for _, peer := range ti.Peers {
			peers = appendWithoutDuplicates(peers, peer)
		}
	}

	trackerInfo.Peers = peers
	return trackerInfo, nil
}

func appendWithoutDuplicates(peers []model.PeerInfo, peer model.PeerInfo) []model.PeerInfo {
	var found bool
	for _, p := range peers {
		if reflect.DeepEqual(p, peer) {
			found = true
			break
		}
	}
	if !found {
		return append(peers, peer)
	}

	return peers
}

func getTrackerInfo(infoHash, peerID [20]byte, announce string, length int64, port uint16) (model.TrackerInfo, error) {
	trackerUrl, err := PrepareTrackerURL(infoHash, peerID, announce, length, port)
	if err != nil {
		return model.TrackerInfo{}, fmt.Errorf("failed to prepare TrackerURL: %w", err)
	}

	resp, err := http.Get(prepareGetRequestUrl(*trackerUrl))
	if err != nil {
		return model.TrackerInfo{}, fmt.Errorf("failed to do get request: %w", err)
	}
	defer func() {
		if resp != nil {
			if errClose := resp.Body.Close(); errClose != nil {
				log.Printf("failed to close resp Body: %w", errClose)
			}
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return model.TrackerInfo{}, fmt.Errorf("bad status: %s", resp.Status)
	}

	return bencode.ParseTrackerInfo(resp.Body)
}

func prepareGetRequestUrl(u url.URL) string {
	return fmt.Sprintf(
		`http://%s%s?%s`,
		u.Host,
		u.Path,
		u.RawQuery,
	)
}

func PrepareTrackerURL(infoHash, peerID [20]byte, announce string, length int64, port uint16) (*url.URL, error) {
	base, err := url.Parse(announce)
	if err != nil {
		return nil, err
	}
	params := url.Values{
		"info_hash":  []string{string(infoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.FormatInt(length, 10)},
	}

	base.RawQuery = params.Encode()
	return base, nil
}
