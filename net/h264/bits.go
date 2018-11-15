package h264

import (
	"errors"
	"io"
)

var (
	errMissingTrailingBits = errors.New("RBSP is missing trailing stop and align bits")
)

// BitReader implements bit stream reading for H.264 network layer processing.
type BitReader struct {
	buffer []byte
	count  uint
	byten  uint
	bitn   uint
}

// NewBitReader creates a bit stream reader from a slice of bytes.
func NewBitReader(buf []byte) *BitReader {
	return &BitReader{buffer: buf, count: uint(len(buf)), byten: 0, bitn: 0}
}

// Available returns number of unread bits in the bit stream buffer.
func (r *BitReader) Available() uint {
	return (r.count-r.byten)*8 - r.bitn
}

// ReadBits attempts to read requested number of bits from the bit stream.
// Returns an error if running into end of stream prematurely.
func (r *BitReader) ReadBits(n uint) (val uint32, err error) {
	if n > 32 || n > r.Available() {
		return 0, io.ErrUnexpectedEOF
	}
	val = uint32(r.buffer[r.byten])
	bits := 8 - r.bitn
	for n > bits {
		// No check here as we already know there is enough data from AvailableBits().
		r.byten++
		val = (val << 8) | uint32(r.buffer[r.byten])
		bits += 8
	}
	// Align and nask out most significant bits that are not requested.
	val = (val >> (bits - n)) & ((1 << n) - 1)
	r.bitn = 8 - (bits - n)
	return
}

// ReadByteBits attempts to read a single byte from the bit stream.
// Returns an error if running into end of stream prematurely.
// Useful to assign result directly to byte target.
func (r *BitReader) ReadByteBits(n uint) (val byte, err error) {
	if n > 8 {
		n = 8
	}
	v, err := r.ReadBits(n)
	return byte(v), err
}

// ReadFlag attempts to read a single bit from the bit stream.
// Returns an error if running into end of stream prematurely.
// Useful to assign result directly to boolean target.
func (r *BitReader) ReadFlag() (val bool, err error) {
	v, err := r.ReadBits(1)
	return v != 0, err
}

// ReadUnsignedGolomb reads and decodes exponential golomb encoded value from a bit stream.
// Returns an error if running into end of stream prematurely.
func (r *BitReader) ReadUnsignedGolomb() (val uint32, err error) {
	var zeroes uint
	for val, err = r.ReadBits(1); val == 0; val, err = r.ReadBits(1) {
		if err != nil {
			// Reached end of stream while all the read bits are zeroes
			return
		}
		zeroes++
	}
	// Last read bit was 1.  Read (zeroes) more.
	val, err = r.ReadBits(zeroes)
	if err != nil {
		return
	}
	// Account for the one extra bit that broke the loop above and subtract 1 per colomb encoding definition
	return val + (1 << zeroes) - 1, nil
}

// ReadSignedGolomb reads and decodes exponential golomb encoded value from a bit stream. and interprets it as signed value.
// Returns an error if running into end of stream prematurely.
func (r *BitReader) ReadSignedGolomb() (val int32, err error) {
	uval, err := r.ReadUnsignedGolomb()
	if err != nil {
		return 0, err
	}
	if (uval & 1) == 0 {
		return -int32(uval / 2), nil
	}
	return int32((uval + 1) / 2), nil
}

// ReadScalingList reads scaling list into a slice and returns either an indicator
// to use default matrix or an error.  The length of the list to read is driven
// by the length of the slice.
func (r *BitReader) ReadScalingList(list []int32) (useDefault bool, err error) {
	lastScale := int32(8)
	nextScale := int32(8)
	var delta int32
	for j := 0; j < len(list); j++ {
		if nextScale != 0 {
			if delta, err = r.ReadSignedGolomb(); err != nil {
				return
			}
			nextScale = (lastScale + delta + 256) % 256
			useDefault = j == 0 && nextScale == 0
		}
		if nextScale == 0 {
			list[j] = lastScale
		} else {
			list[j] = nextScale
			lastScale = nextScale
		}
	}
	return
}

// SkipBits attempts to skip requested number of bits in the bit stream.
// Returns an error if running into end of stream prematurely.
func (r *BitReader) SkipBits(n uint) (err error) {
	if n > 32 || n > r.Available() {
		return io.ErrUnexpectedEOF
	}
	bits := 8 - r.bitn
	for n > bits {
		// No check here as we already know there is enough data from AvailableBits().
		r.byten++
		bits += 8
	}
	r.bitn = 8 - (bits - n)
	return
}

// SkipGolomb decodes exponential golomb encoded value in the bit stream and skips it.
// Returns an error if running into end of stream prematurely.
func (r *BitReader) SkipGolomb() (err error) {
	var zeroes uint
	var val uint32
	for val, err = r.ReadBits(1); val == 0; val, err = r.ReadBits(1) {
		if err != nil {
			// Reached end of stream while all the read bits are zeroes
			return
		}
		zeroes++
	}
	// Last read bit was 1.  Read (zeroes) more.
	return r.SkipBits(zeroes)
}

// SkipTrailingBits ensures that remaining bits in the buffer consist of stop bit and
// variable number of aligning zero-bits;
func (r *BitReader) SkipTrailingBits() (err error) {
	v, err := r.ReadBits(1)
	if err != nil || v == 0 {
		return errMissingTrailingBits
	}
	for r.bitn != 0 && r.bitn != 8 {
		v, err := r.ReadBits(1)
		if err != nil || v != 0 {
			return errMissingTrailingBits
		}
	}
	return nil
}

// ReadPayloadParam parses payload type and size of SEI message from bit reader.
func (r *BitReader) ReadPayloadParam() (val uint32, err error) {
	var v uint32
	for {
		if v, err = r.ReadBits(8); err != nil {
			return
		}
		val += v
		if v != 0xff {
			return
		}
	}
}
