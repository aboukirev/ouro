package rtp

import (
	"encoding/binary"
	"errors"
)

const (
	// RtpVersion is used to verify compliance with current specification of the RTP protocol.
	RtpVersion = 2 << 6
	// HeaderSize defines the size of the fixed part of the packet, up to and inclding SSRC.
	HeaderSize = 12
)

var (
	errInvalidVersion = errors.New("Invalid version of RTP packet")
)

var be = binary.BigEndian

type (
	// Packet encapsulates RTP packet structure.
	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |V=2|P|X|  CC   |M|     PT      |       sequence number         |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                           timestamp                           |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |           synchronization source (SSRC) identifier            |
	// +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
	// |            contributing source (CSRC) identifiers             |
	// |                             ....                              |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	Packet struct {
		VPXCC byte     // Version, Padding, Extension, Contributing Source Count
		MPT   byte     // Marker, Payload Type
		SN    uint16   // Sequense Number
		TS    uint32   // Timestamp
		SSRC  uint32   // Synchronization Source Identifier
		CSRC  []uint32 // Contributing Source Identifiers
		XH    uint16   // Extension Header (profile dependent)
		XL    uint16   // Extension Length (in `uint`s not inclusing this header)
		XD    []byte   // Extension Data
		PL    []byte   // Payload
	}
)

// Unpack validates a packed RTP packet and converts it into a sparse structure.
func Unpack(buf []byte) (*Packet, error) {
	if (buf[0] & 0xC0) != RtpVersion {
		return nil, errInvalidVersion
	}
	packet := &Packet{
		VPXCC: buf[0],
		MPT:   buf[1],
		SN:    be.Uint16(buf[2:]),
		TS:    be.Uint32(buf[4:]),
		SSRC:  be.Uint32(buf[8:]),
	}

	off := HeaderSize
	packet.CSRC = make([]uint32, packet.CC())
	for i := range packet.CSRC {
		packet.CSRC[i] = be.Uint32(buf[off:])
		off += 4
	}

	if packet.X() {
		packet.XH = be.Uint16(buf[off:])
		packet.XL = be.Uint16(buf[off+2:])
		off += 4
		if packet.XL > 0 {
			packet.XD = buf[off : off+int(packet.XL)*4]
			off += int(packet.XL) * 4
		}
	}

	packet.PL = buf[off:]

	//s.rtpChan <- packet
	return packet, nil
}

// P returns Padding flag value of the packet.
func (p Packet) P() bool {
	return (p.VPXCC & 0x20) != 0
}

// X returns Extension flag value of the packet.
func (p Packet) X() bool {
	return (p.VPXCC & 0x10) != 0
}

// CC returns Contributing Source Count of the packet.
func (p Packet) CC() byte {
	return p.VPXCC & 0x0F
}

// M returns Marker value of the packet.
func (p Packet) M() bool {
	return (p.MPT & 0x80) != 0
}

// PT returns Payload Type of the packet.
func (p Packet) PT() byte {
	return p.MPT & 0x7F
}

// Pack converts sparse RTP packet into a slice of bytes for network.
func (p *Packet) Pack() []byte {
	if p == nil {
		return nil
	}
	sz := uint16(HeaderSize + p.CC()*4)
	if p.X() {
		sz = sz + 4*(1+p.XL)
	}
	b := make([]byte, sz)
	b[0] = p.VPXCC
	b[1] = p.MPT
	be.PutUint16(b[2:], p.SN)
	be.PutUint32(b[4:], p.TS)
	// TODO: Finish this
	return b
}
