package bencode

type (
	bitTorrent struct {
		Announce     string     `bencode:"announce"`
		AnnounceList [][]string `bencode:"announce-list"`
		Comment      string     `bencode:"comment"`
		CreatedBy    string     `bencode:"created by"`
		CreationDate int64      `bencode:"creation date"`
		Encoding     string     `bencode:"encoding"`
		Info         info       `bencode:"info"`
	}

	info struct {
		Pieces      string `bencode:"pieces"`
		PieceLength int64  `bencode:"piece length"`
		Length      int64  `bencode:"length"`
		Name        string `bencode:"name"`
	}

	trackerResponse struct {
		Interval int64  `bencode:"interval"`
		Peers    string `bencode:"peers"`
	}
)
