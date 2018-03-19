package sdp

// TODO: A more comprehensive parsing.

import (
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
)

// Codecs
const (
	H264 = 96
	AAC  = 97
)

type (
	// Media represents descriptor for media stream supported by RTSP source.
	Media struct {
		audio              bool
		Type               uint
		TimeScale          int
		Control            string
		Rtpmap             int
		Config             []byte
		SpropParameterSets [][]byte
		PayloadType        int
		SizeLength         int
		IndexLength        int
	}
)

// Parse parses body of RTSP response to DESCRIBE command and returns information about playable media feeds.
func Parse(buf []byte) (feeds []Media) {
	var m *Media
	content := string(buf)

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		keyval := strings.SplitN(line, "=", 2)
		if len(keyval) == 2 {
			fields := strings.SplitN(keyval[1], " ", 2)

			switch keyval[0] {
			case "m":
				if len(fields) > 0 {
					switch fields[0] {
					case "video": // case "audio", "video":
						feeds = append(feeds, Media{audio: fields[0] == "audio"})
						m = &feeds[len(feeds)-1]
						mfields := strings.Split(fields[1], " ")
						if len(mfields) >= 3 {
							m.PayloadType, _ = strconv.Atoi(mfields[2])
						}
					default:
						m = nil
					}
				}
			case "a":
				if m != nil {
					for _, field := range fields {
						keyval = strings.SplitN(field, ":", 2)
						if len(keyval) >= 2 {
							key := keyval[0]
							val := keyval[1]
							switch key {
							case "control":
								m.Control = val
							case "rtpmap":
								m.Rtpmap, _ = strconv.Atoi(val)
							}
						}
						keyval = strings.Split(field, "/")
						if len(keyval) >= 2 {
							key := keyval[0]
							switch strings.ToUpper(key) {
							case "MPEG4-GENERIC":
								m.Type = AAC
							case "H264":
								m.Type = H264
							}
							if i, err := strconv.Atoi(keyval[1]); err == nil {
								m.TimeScale = i
							}
						}
						keyval = strings.Split(field, ";")
						if len(keyval) > 1 {
							for _, field := range keyval {
								keyval := strings.SplitN(field, "=", 2)
								if len(keyval) == 2 {
									key := strings.TrimSpace(keyval[0])
									val := keyval[1]
									switch key {
									case "config":
										m.Config, _ = hex.DecodeString(val)
									case "sizelength":
										m.SizeLength, _ = strconv.Atoi(val)
									case "indexlength":
										m.IndexLength, _ = strconv.Atoi(val)
									case "sprop-parameter-sets":
										fields := strings.Split(val, ",")
										for _, field := range fields {
											val, _ := base64.StdEncoding.DecodeString(field)
											m.SpropParameterSets = append(m.SpropParameterSets, val)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return
}
