package bencode

import (
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"torrent-server/model"
)

var (
	correctText        = "d8:announce12:testAnnounce13:announce-listll13:testAnnounce1el13:testAnnounce2ee7:comment11:testComment10:created by14:uTorrent/3.5.513:creation datei1609502400e8:encoding5:UTF-84:infod6:lengthi835109565e4:name8:testName12:piece lengthi1048576e6:pieces20:testPiecesTestPiecesee"
	expectedBitTorrent = model.BitTorrent{
		Announce: "testAnnounce",
		AnnounceList: append(make([][]string, 0, 8),
			append(make([]string, 0, 8), "testAnnounce1"),
			append(make([]string, 0, 8), "testAnnounce2"),
		),
		Comment:      "testComment",
		CreatedBy:    "uTorrent/3.5.5",
		CreationDate: time.Unix(1609502400, 0),
		Encoding:     "UTF-8",
		InfoHash:     [20]byte{6, 127, 105, 243, 13, 160, 43, 122, 48, 152, 63, 198, 157, 143, 99, 9, 234, 213, 2, 250},
		PieceHashes: [][20]byte{
			{116, 101, 115, 116, 80, 105, 101, 99, 101, 115, 84, 101, 115, 116, 80, 105, 101, 99, 101, 115},
		},
		PieceLength: 1048576,
		Length:      835109565,
		Name:        "testName",
	}

	corruptedText = "corrupted"
)

func TestBenCodeParserParse(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    model.BitTorrent
		wantErr bool
	}{
		{
			name: "correct",
			args: args{r: strings.NewReader(correctText)},
			want: expectedBitTorrent,
		},
		{
			name:    "corrupted",
			args:    args{r: strings.NewReader(corruptedText)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			be := &benCodeParser{}
			got, err := be.Parse(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	bParser := NewParser()
	reader := strings.NewReader(correctText)
	for i := 0; i < b.N; i++ {
		_, _ = bParser.Parse(reader)
	}
}
