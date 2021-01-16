package bencode

import "torrent-server/model"

func toDomainModel(torrent bitTorrent) model.BitTorrent {
	return model.BitTorrent{
		Announce:     torrent.Announce,
		//AnnounceList: torrent.AnnounceList,
		Comment:      torrent.Comment,
		Info: model.Info{
			Pieces:      torrent.Info.Pieces,
			//PieceLength: torrent.Info.PieceLength,
			//Length:      torrent.Info.Length,
			Name:        torrent.Info.Name,
		},
	}
}
