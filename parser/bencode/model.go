package bencode

type (
	bitTorrent struct {
		Announce     string   `bencode:"announce"`
		AnnounceList [][]string `bencode:"announce-list"`
		Comment      string   `bencode:"comment"`
		CreatedBy    string   `bencode:"created by"`
		CreationDate uint      `bencode:"creation date"`
		Encoding     string   `bencode:"encoding"`
		Info         info     `bencode:"info"`
	}

	info struct {
		Pieces      string `bencode:"pieces"`
		PieceLength uint    `bencode:"piece length"`
		Length      uint    `bencode:"length"`
		Name        string `bencode:"name"`
	}
)
