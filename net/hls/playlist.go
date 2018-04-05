package hls

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

const (
	// SegmentTypeMPEGTS identifies segment contents as MPEG-2 Transport Stream
	SegmentTypeMPEGTS = iota
	// SegmentTypeFMP4 identifies segment contents as Fragmented MPEG-4
	SegmentTypeFMP4
)

type (
	// Segment describes a chunk of streaming media.
	Segment struct {
		Duration float64
		Position int64
		Length   int64
	}

	// Playlist encapsulates HLS playlist.
	//
	// The underlying media source is created by other code.  It is just a file for the playlist.
	// A configurable number of segments is organized into a ring and constitutes a series of byte ranges in the file.
	Playlist struct {
		Version  int       // Protocol version.  We need at least 4 for EXT-X-BYTERANGE.
		FileName string    // Stream source: either MPEG-TS or FMP4.
		URI      string    // URI at which media file is served.
		Segments []Segment // Acts as a ring of segments with wrap over.
		SegSize  float64   // Desired segment size.
		First    int       // First segment in the ring.
		Last     int       // Last segment in the ring.
	}
)

// NewPlaylist verifies that file exists and accessible for reading then instantiates Playlist.
func NewPlaylist(uri string, name string, nseg int, size float64) (*Playlist, error) {
	f, err := os.OpenFile(name, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	base := filepath.Base(name)
	return &Playlist{
		Version:  4,
		URI:      path.Join(uri, base),
		FileName: name,
		Segments: make([]Segment, nseg),
		SegSize:  size,
		First:    -1,
		Last:     -1,
	}, nil
}

// String renders current state of Playlist as M3U8.
func (pl Playlist) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("#EXTM3U\n")
	buf.WriteString("#EXT-X-VERSION:")
	buf.WriteString(strconv.Itoa(pl.Version))
	buf.WriteByte('\n')
	if pl.First >= 0 {
		nseg := len(pl.Segments)
		for i := pl.First; ; i = (i + 1) % nseg {
			seg := pl.Segments[i]
			buf.WriteString("#EXTINF:")
			buf.WriteString(strconv.FormatFloat(seg.Duration, 'f', -1, 64))
			buf.WriteByte('\n')
			buf.WriteString("#EXT-X-BYTERANGE:")
			buf.WriteString(strconv.FormatInt(seg.Length, 10))
			buf.WriteByte('@')
			buf.WriteString(strconv.FormatInt(seg.Position, 10))
			buf.WriteByte('\n')
			buf.WriteString(pl.URI)
			buf.WriteByte('\n')

			if i == pl.Last {
				break
			}
		}
	}
	buf.WriteString("#EXT-X-ENDLIST\n")
	return buf.String()
}

// AddSegment advances to next position in the ring and populates segment structure.
// Position is calculates from previous segment or is 0 at the start.
func (pl *Playlist) AddSegment(duration float64, length int64) {
	pos := int64(0)
	if pl.Last >= 0 {
		pos = pl.Segments[pl.Last].Position + pl.Segments[pl.Last].Length
	}
	pl.Last = (pl.Last + 1) % len(pl.Segments)
	pl.Segments[pl.Last].Duration = duration
	pl.Segments[pl.Last].Position = pos
	pl.Segments[pl.Last].Length = length
	if pl.First == pl.Last {
		pl.First++
	} else if pl.First < 0 {
		pl.First = pl.Last
	}
}
