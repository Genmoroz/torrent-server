package bencode

import (
	"strings"
	"testing"
)

var test = "d8:announce12:testAnnounce13:announce-listll13:testAnnounce1el13:testAnnounce2ee7:comment11:testComment10:created by14:uTorrent/3.5.513:creation datei1610722879e8:encoding5:UTF-84:infod6:lengthi835109565e4:name8:testName12:piece lengthi1048576e6:pieces10:testPiecesee"

//func Test_benCodeParser_Parse(t *testing.T) {
//	type args struct {
//		r io.Reader
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    model.BitTorrent
//		wantErr bool
//	}{
//		{
//			name: "correct",
//			args: args{r: strings.NewReader(test)},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			be := &benCodeParser{}
//			got, err := be.Parse(tt.args.r)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("Parse() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func BenchmarkParse(b *testing.B) {
	parser := NewParser()
	reader := strings.NewReader(test)
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(reader)
	}
}
