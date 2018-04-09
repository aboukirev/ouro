package hls

import (
	"testing"
)

func TestPlaylist(t *testing.T) {
	list := `#EXTM3U
#EXT-X-VERSION:4
#EXTINF:3.014
#EXT-X-BYTERANGE:96448@82560
https://example.com/media/video.ts
#EXTINF:3.24
#EXT-X-BYTERANGE:103680@179008
https://example.com/media/video.ts
#EXTINF:2.9777
#EXT-X-BYTERANGE:95286@282688
https://example.com/media/video.ts
#EXTINF:3.4333
#EXT-X-BYTERANGE:109866@377974
https://example.com/media/video.ts
#EXTINF:3.41
#EXT-X-BYTERANGE:109120@487840
https://example.com/media/video.ts
#EXT-X-ENDLIST
`
	segments := []struct {
		duration float64
		length   int64
	}{
		{2.58, 82560},
		{3.014, 96448},
		{3.24, 103680},
		{2.9777, 95286},
		{3.4333, 109866},
		{3.41, 109120},
	}
	pl, err := NewPlaylist("https://example.com/media", "/var/media/video.ts", 5, 3.5)
	if err != nil {
		t.Fatal(err)
	}
	for _, seg := range segments {
		pl.AddSegment(seg.duration, seg.length)
	}
	if pl.String() != list {
		t.Error(pl)
	}
}
