package bencode

type (
	bitTorrent struct {
		Announce     string     `bencode:"announce"`
		AnnounceList [][]string `bencode:"announce-list"`
		Comment      string     `bencode:"comment"`
		CreatedBy    string     `bencode:"created by"`
		CreationDate uint32     `bencode:"creation date"`
		Encoding     string     `bencode:"encoding"`
		Info         info       `bencode:"info"`
	}

	info struct {
		Pieces      string `bencode:"pieces"`
		PieceLength uint32 `bencode:"piece length"`
		Length      uint32 `bencode:"length"`
		Name        string `bencode:"name"`
	}

	trackerResponse struct {
		Interval uint32 `bencode:"interval"`
		Peers    string `bencode:"peers"`
	}
)
