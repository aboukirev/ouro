package rtsp

// There are 3 modes of communication:
// - if lower protocol is TCP then it is RTP over RTSP, i.e. RTP packets are interleaved and sent over RTSP.
// - if lower protocol is unicast UDP then we preallocate a pair of UDP sockets for each media to listen on and provide them as client ports in transport setup.
//   First port in pair is for data, second is for control (RTCP).
// - if lower protocol is multicast then we just wait for response to transport setup and get client ports from there for each media.
// Data port should be even, control port - odd.
// Looks like surveillance cameras hate multicast, it is mostly for online video conferencing.  Or I'll have to figure out why I am getting 400 "Bad request".

import (
	"errors"
	"strconv"
	"strings"
)

var (
	// ErrMalformedTransport indicates trouble parsing Transport header value.
	ErrMalformedTransport = errors.New("Malformed value of Transport header")
)

const (
	// ProtoTCP requests TCP as lower protocol.
	ProtoTCP = 0
	// ProtoUnicast requests UDP unicast as lower protocol.
	ProtoUnicast = 1
	// ProtoMulticast requests UDP multicast as lower protocol.
	ProtoMulticast = 2
	// ProtoHTTP opens HTTP connections with GET and POST, transmits and receives base64 encoded RTSP messages over it.
	ProtoHTTP = 3
)

type (
	// Pair represents a pair of channels or ports.
	Pair struct {
		One int
		Two int
	}

	// Transport encapsulates RTSP transport header information.
	Transport struct {
		IsTCP         bool
		IsMulticast   bool
		IsInterleaved bool
		IsAppend      bool
		Interleave    Pair
		Port          Pair
		ClientPort    Pair
		ServerPort    Pair
		Layers        int
		TTL           int
		Destination   string
		Source        string
		SSRC          string
		Mode          string
	}
)

// Transport           =   "Transport" ":"
//                         1\#transport-spec
// transport-spec      =   transport-protocol/profile[/lower-transport]
//                         *parameter
// transport-protocol  =   "RTP"
// profile             =   "AVP"
// lower-transport     =   "TCP" | "UDP"
// parameter           =   ( "unicast" | "multicast" )
//                     |   ";" "destination" [ "=" address ]
//                     |   ";" "interleaved" "=" channel [ "-" channel ]
//                     |   ";" "append"
//                     |   ";" "ttl" "=" ttl
//                     |   ";" "layers" "=" 1*DIGIT
//                     |   ";" "port" "=" port [ "-" port ]
//                     |   ";" "client_port" "=" port [ "-" port ]
//                     |   ";" "server_port" "=" port [ "-" port ]
//                     |   ";" "ssrc" "=" ssrc
//                     |   ";" "mode" = <"> 1\#mode <">
// ttl                 =   1*3(DIGIT)
// port                =   1*5(DIGIT)
// ssrc                =   8*8(HEX)
// channel             =   1*3(DIGIT)
// address             =   host
// mode                =   <"> *Method <"> | Method

// NewTransport creates default transport for media.
func NewTransport(proto int, port int) *Transport {
	t := &Transport{IsAppend: false, Mode: "PLAY"}
	switch proto {
	case ProtoTCP, ProtoHTTP:
		t.IsTCP = true
		t.IsInterleaved = true
		t.Interleave = Pair{One: port, Two: port + 1}
	case ProtoUnicast:
		t.ClientPort = Pair{One: port, Two: port + 1}
	case ProtoMulticast:
		t.IsMulticast = true
		t.Port = Pair{One: port, Two: port + 1}
	}
	return t
}

// String formats transport parameters into a value for respective RTSP header.
func (t Transport) String() string {
	b := strings.Builder{}
	b.WriteString("RTP/AVP")
	if t.IsTCP {
		b.WriteString("/TCP")
	} else {
		//b.WriteString("/UDP")   // UDP assumed by default.
		if t.IsMulticast {
			b.WriteString(";multicast")
		} else {
			b.WriteString(";unicast")
		}
	}
	if t.Destination != "" {
		b.WriteString(";destination=")
		b.WriteString(t.Destination)
	}
	if t.Source != "" {
		b.WriteString(";source=")
		b.WriteString(t.Source)
	}
	if t.IsInterleaved {
		b.WriteString(";interleaved=")
		b.WriteString(t.Interleave.String())
	}
	if t.IsAppend {
		b.WriteString(";append")
	}
	if t.Port.One > 0 {
		b.WriteString(";port=")
		b.WriteString(t.Port.String())
	}
	if t.ClientPort.One > 0 {
		b.WriteString(";client_port=")
		b.WriteString(t.ClientPort.String())
	}
	if t.ServerPort.One > 0 {
		b.WriteString(";server_port=")
		b.WriteString(t.ServerPort.String())
	}
	// TTL parameter applies only for UDP multicast.
	if t.IsMulticast && t.TTL > 0 {
		b.WriteString(";ttl=")
		b.WriteString(strconv.Itoa(t.TTL))
	}
	if t.Layers > 0 {
		b.WriteString(";layers=")
		b.WriteString(strconv.Itoa(t.Layers))
	}
	if t.SSRC != "" {
		b.WriteString(";ssrc=")
		b.WriteString(t.SSRC)
	}
	if t.Mode != "" && t.Mode != "PLAY" {
		b.WriteString(";mode=")
		b.WriteString(t.Mode)
	}
	return b.String()
}

// Parse transport header into constituent parts.
func (t *Transport) Parse(value string) (err error) {
	fields := strings.Split(value, ";")
	for _, field := range fields {
		keyval := strings.Split(field, "/")
		if len(keyval) == 3 {
			// Third part is lower protocol TCP/UDP
			t.IsTCP = strings.ToUpper(keyval[2]) == "TCP"
		}
		keyval = strings.SplitN(field, "=", 2)
		switch strings.ToLower(keyval[0]) {
		case "unicast":
			t.IsMulticast = false
		case "multicast":
			t.IsMulticast = true
		case "destination":
			if len(keyval) == 2 {
				t.Destination = keyval[1]
			} else {
				err = ErrMalformedTransport
			}
		case "source":
			if len(keyval) == 2 {
				t.Source = keyval[1]
			} else {
				err = ErrMalformedTransport
			}
		case "interleaved":
			t.IsMulticast = false
			if len(keyval) > 1 {
				t.Interleave, err = ParsePair(keyval[1])
			} else {
				err = ErrMalformedTransport
			}
		case "append":
			t.IsAppend = true
		case "layers":
			if len(keyval) > 1 {
				t.Layers, err = strconv.Atoi(keyval[1])
			} else {
				err = ErrMalformedTransport
			}
		case "ttl":
			if len(keyval) > 1 {
				t.TTL, err = strconv.Atoi(keyval[1])
			} else {
				err = ErrMalformedTransport
			}
		case "port":
			if len(keyval) > 1 {
				t.Port, err = ParsePair(keyval[1])
			} else {
				err = ErrMalformedTransport
			}
		case "client_port":
			if len(keyval) > 1 {
				t.ClientPort, err = ParsePair(keyval[1])
			} else {
				err = ErrMalformedTransport
			}
		case "server_port":
			if len(keyval) > 1 {
				t.ServerPort, err = ParsePair(keyval[1])
			} else {
				err = ErrMalformedTransport
			}
		case "ssrc":
			if len(keyval) > 1 {
				t.SSRC = keyval[1]
			} else {
				err = ErrMalformedTransport
			}
		case "mode":
			if len(keyval) > 1 {
				t.Mode = keyval[1]
			} else {
				err = ErrMalformedTransport
			}
		}
		if err != nil {
			return
		}
	}
	return
}

// ParsePair parses a pair of channels or ports in transport header.
func ParsePair(val string) (p Pair, err error) {
	parts := strings.Split(val, "-")
	p.One, err = strconv.Atoi(parts[0])
	if err == nil && len(parts) > 1 {
		p.Two, err = strconv.Atoi(parts[1])
		if err != nil {
			p.Two = p.One + 1
		}
	}
	return
}

// String formats a pair of integers into channels or ports in transport header.
func (p Pair) String() string {
	return strconv.Itoa(p.One) + "-" + strconv.Itoa(p.Two)
}
