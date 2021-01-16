package bencode

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/jackpal/bencode-go"
	"time"

	"torrent-server/model"
)

func toDomainModel(torrent bitTorrent) (model.BitTorrent, error) {
	pieceHashes, err := splitPieceHashes(torrent.Info.Pieces)
	if err != nil {
		return model.BitTorrent{}, fmt.Errorf("failed to split piece hashes: %w", err)
	}

	infoHash, err := infoHash(torrent.Info)

	return model.BitTorrent{
		Announce:     torrent.Announce,
		AnnounceList: torrent.AnnounceList,
		Comment:      torrent.Comment,
		CreatedBy:    torrent.CreatedBy,
		CreationDate: time.Unix(torrent.CreationDate, 0),
		Encoding:     torrent.Encoding,
		InfoHash:     infoHash,
		PieceHashes:  pieceHashes,
		PieceLength:  torrent.Info.PieceLength,
		Length:       torrent.Info.Length,
		Name:         torrent.Info.Name,
	}, nil
}

func splitPieceHashes(s string) ([][20]byte, error) {
	hashLen := 20 // Length of SHA-1 hash

	buf := []byte(s)
	if len(buf)%hashLen != 0 {
		return nil, fmt.Errorf("received malformed pieces of length %d", len(buf))
	}

	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)
	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}

	return hashes, nil
}

func infoHash(i info) ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, i)
	if err != nil {
		return [20]byte{}, fmt.Errorf("failed to marshal: %w", err)
	}

	return sha1.Sum(buf.Bytes()), nil
}
