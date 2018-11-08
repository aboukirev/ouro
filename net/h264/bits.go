package h264

import (
	"io"
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

// Read attempts to read requested number of bits from the bit stream.
// Returns an error if running into end of stream prematurely.
func (r *BitReader) Read(n uint) (val uint32, err error) {
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

// ReadExponentialGolomb reads and decodes exponential golomb encoded value from a bit stream.
// Returns an error if running into end of stream prematurely.
func (r *BitReader) ReadExponentialGolomb() (val uint32, err error) {
	var zeroes uint
	for val, err = r.Read(1); val == 0; val, err = r.Read(1) {
		if err != nil {
			// Reached end of stream while all the read bits are zeroes
			return
		}
		zeroes++
	}
	// Last read bit was 1.  Read (zeroes) more.
	val, err = r.Read(zeroes)
	if err != nil {
		return
	}
	// Account for the one extra bit that broke the loop above and subtract 1 per colomb encoding definition
	return val + (1 << zeroes) - 1, nil
}

// ReadSignedGolomb reads and decodes exponential golomb encoded value from a bit stream. and interprets it as signed value.
// Returns an error if running into end of stream prematurely.
func (r *BitReader) ReadSignedGolomb() (val int32, err error) {
	uval, err := r.ReadExponentialGolomb()
	if err != nil {
		return 0, err
	}
	if (uval & 1) == 0 {
		return -int32(uval / 2), nil
	}
	return int32((uval + 1) / 2), nil
}
