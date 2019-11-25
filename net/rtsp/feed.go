package rtsp

import (
	"github.com/daneshvar/ouro/net/h264"
	"github.com/daneshvar/ouro/net/sdp"
)

type (
	// Feed represents a media feed encapsulating SDP, transport properties,
	// channel, and utility functions.
	Feed struct {
		sdp.Media
		transp *Transport
		cseq   int
		IsSet  bool
		Sets   *h264.ParameterSets
	}
)

// ParseFeeds creates media feeds from parsed SDP body.
func ParseFeeds(proto int, buf []byte) (feeds []*Feed, err error) {
	for i, m := range sdp.Parse(buf) {
		feeds = append(feeds, &Feed{Media: m, transp: NewTransport(proto, i*2), Sets: h264.NewParameterSets()})
	}
	return
}

// TransportHeader returns a formatted transport header for SETUP request.
func (f *Feed) TransportHeader() string {
	return f.transp.String()
}

// TransportSetup populates transport properties from response to SETUP verb.
func (f *Feed) TransportSetup(value string) error {
	if err := f.transp.Parse(value); err != nil {
		return err
	}
	// Parse sprop parameter sets from the SDP.
	for _, b := range f.SpropParameterSets {
		if err := f.Sets.ParseSprop(b); err != nil {
			return err
		}
	}
	f.IsSet = true
	return nil
}
