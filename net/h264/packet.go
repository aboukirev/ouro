package h264

import (
	"encoding/binary"
	"errors"
)

// TODO: Parse Annex B NAL sequences.  Currently code assumes raw NALs where length of payload is declared in container.

// Type   Packet     Type name
// ------------------------------------------------
// 0      undefined  -
// 1      Non-IDR coded slice
// 2      Data Partition A
// 3      Data Partition B
// 4      Data Partition C
// 5      IDR (Instantaneous Decoding Refresh) Picture
// 6      SEI (Supplemental Enhancement Information)
// 7      SPS (Sequence Parameter Set)
// 8      PPS (Picture Parameter Set)
// 9      Access Unit Delimiter
// 10     EoS (End of Sequence)
// 11     EoS (End of Stream)
// 12     Filler Data
// ...
// 24     STAP-A     Single-time aggregation packet
// 25     STAP-B     Single-time aggregation packet
// 26     MTAP16     Multi-time aggregation packet
// 27     MTAP24     Multi-time aggregation packet
// 28     FU-A       Fragmentation unit
// 29     FU-B       Fragmentation unit
const (
	typeNIDR   = 1
	typeDPA    = 2
	typeDPB    = 3
	typeDPC    = 4
	typeIDR    = 5
	typeSEI    = 6
	typeSPS    = 7
	typePPS    = 8
	typeAUD    = 9
	typeEOSeq  = 10
	typeEOStr  = 11
	typeFill   = 12
	typeStapA  = 24
	typeStapB  = 25
	typeMtap16 = 26
	typeMtap24 = 27
	typeFuA    = 28
	typeFuB    = 29
)

type (
	// NALUnit describes a single Network Access Layer Unit in video stream.
	// +---------------+
	// |0|1|2|3|4|5|6|7|
	// +-+-+-+-+-+-+-+-+
	// |F|NRI| Type    |
	// +---------------+
	// Forbidden Zero bit, must be 0; NAL Ref Idc; NAL Type.
	NALUnit struct {
		Header byte
		Don    uint16 // Decoding Order Number
		TS     uint32 // Timestamp
		Data   []byte
	}

	// NALFragment represents NAL unit fragment together with fragment specidfic flags (start, end).
	// +---------------+
	// |0|1|2|3|4|5|6|7|
	// +-+-+-+-+-+-+-+-+
	// |S|E|R| Type    |
	// +---------------+
	// Start flag, End flag, Reserved must be 0, NAL payload Type
	NALFragment struct {
		NALUnit
		Flags byte
	}

	// NALSink handles NAL unit aggreagates and fragments
	NALSink struct {
		Units     []NALUnit
		Fragments []NALFragment
		Don       uint16 // Decoding Order Number
	}
)

var (
	errNeedPacket           = errors.New("Allocate packet to parse incoming data into")
	errInvalidPayloadHeader = errors.New("Invalid or malformed payload header")
	errPacketTooShort       = errors.New("Packet is too short")
)

var be = binary.BigEndian

// ZeroBit returns forbidden zero bit value of NAL header.
func (u *NALUnit) ZeroBit() bool {
	return (u.Header & 0x80) == 0
}

// RefIdc returns NAL unit reference flag and importance.
func (u *NALUnit) RefIdc() byte {
	return (u.Header & 0x60) >> 5
}

// Type returns type of NAL unit.
func (u *NALUnit) Type() byte {
	return (u.Header & 0x1F)
}

// IsStart returns start flag of the NAL unit fragment.
func (f NALFragment) IsStart() bool {
	return (f.Flags & 0x80) != 0
}

// IsEnd returns end flag of the NAL unit fragment.
func (f NALFragment) IsEnd() bool {
	return (f.Flags & 0x40) != 0
}

// NewNALSink creates a sink to handle NAL unit aggreagates and fragments.
// Sink combines fragments to emit a unit into the queue and resets unit queue
// on each subsequent RTP packet.  Fragment queue is reset upon receiving first
// fragment in a series.
func NewNALSink() *NALSink {
	return &NALSink{
		Units:     make([]NALUnit, 0, 20),
		Fragments: make([]NALFragment, 0, 20),
	}
}

// AddFragment pushes NAL unit fragment into the fragment queue, resetting queue before
// accepting first fragment in a series.  Upon receiving last fragment in a series, all fragments
// are combined into a NAL unit and added to the unit queue.
func (s *NALSink) AddFragment(indicator byte, header byte, don uint16, ts uint32, data []byte) {
	if (header & 0x80) != 0 {
		s.Fragments = s.Fragments[:0]
	}
	s.Fragments = append(s.Fragments, NALFragment{
		NALUnit: NALUnit{
			Header: indicator,
			Don:    don,
			TS:     ts,
			Data:   data,
		},
		Flags: header,
	})
	if (header & 0x40) != 0 {
		// TODO: Sort fragments by TS, Don (TBD).
		// Combine fragments and append complete NAL.
		data := []byte{}
		for _, frag := range s.Fragments {
			data = append(data, frag.Data...)
		}
		s.Units = append(s.Units, NALUnit{Header: indicator, Don: don, TS: ts, Data: data})
	}
}

func (s *NALSink) parseSTAP(typ byte, buf []byte, ts uint32) error {
	if typ == typeStapB {
		if len(buf) < 2 {
			return errPacketTooShort
		}
		s.Don = be.Uint16(buf)
		buf = buf[2:]
	}
	for {
		if len(buf) < 2 {
			return nil // Nothing more to process
		}
		size := be.Uint16(buf)
		if len(buf) < int(size+2) {
			return errPacketTooShort
		}
		s.Units = append(s.Units, NALUnit{Header: buf[2], Don: s.Don, TS: ts, Data: buf[3 : size+2]})
		s.Don++
		buf = buf[size+2:]
	}
}

func (s *NALSink) parseMTAP(typ byte, buf []byte, ts uint32) error {
	if len(buf) < 2 {
		return errPacketTooShort
	}
	s.Don = be.Uint16(buf)
	buf = buf[2:]
	off := uint16(2 + 1 + 2) // size, donb, 16-bit ts offset
	if typ == typeMtap24 {
		off++ // 24-bit ts offset instead of 16-bit
	}
	for {
		if len(buf) < int(off) {
			return nil // Nothing more to process
		}
		size := be.Uint16(buf)
		dond := uint16(buf[2])
		// Read and handle 16-bit or 24-bit timestamp offset.
		tsoff := uint32(buf[3])<<8 + uint32(buf[4])
		if typ == typeMtap24 {
			tsoff = ts<<8 + uint32(buf[5])
		}
		if len(buf) < int(size+off) {
			return errPacketTooShort
		}
		s.Units = append(s.Units, NALUnit{Header: buf[off], Don: s.Don + dond, TS: ts + tsoff, Data: buf[off+1 : size+off]})
		s.Don++
		buf = buf[size+off:]
	}
}

// Push RTP payload parsing NAL units and handling aggregation and fragmenting.
func (s *NALSink) Push(buf []byte, ts uint32) error {
	// TODO: Detect Annex B vs AVC payloads.  If starts with Annex B start code then split on start code.
	for _, nal := range splitAnnexB(buf) {
		if err := s.parseNAL(nal, ts); err != nil {
			return err
		}
	}
	return nil
}

func (s *NALSink) parseNAL(buf []byte, ts uint32) error {
	s.Units = s.Units[:0]
	if len(buf) < 1 {
		return errPacketTooShort
	}

	typ := buf[0] & 0x1F
	switch typ {
	case typeStapA:
		return s.parseSTAP(typ, buf[1:], ts)
	case typeStapB:
		return s.parseSTAP(typ, buf[1:], ts)
	case typeMtap16:
		return s.parseMTAP(typ, buf[1:], ts)
	case typeMtap24:
		return s.parseMTAP(typ, buf[1:], ts)
	case typeFuA:
		nri := buf[0] & 0x60
		if len(buf) < 2 {
			return errPacketTooShort
		}
		s.AddFragment(nri+(buf[1]&0x1F), buf[1], s.Don, ts, buf[2:])
	case typeFuB:
		nri := buf[0] & 0x60
		if len(buf) < 4 {
			return errPacketTooShort
		}
		s.AddFragment(nri+(buf[1]&0x1F), buf[1], be.Uint16(buf[2:]), ts, buf[4:])
	default:
		s.Units = append(s.Units, NALUnit{Header: buf[0], Don: s.Don, TS: ts, Data: buf[1:]})
		s.Don++
	}
	return nil
}

// DonDiff may not be needed as uint16 automatically wraps on under/overflow.
func DonDiff(don1 uint16, don2 uint16) (diff int) {
	diff = int(don2) - int(don1)
	if diff >= 32768 {
		diff = diff - 65536
	} else if diff <= -32768 {
		diff = diff + 65536
	}
	return
}
