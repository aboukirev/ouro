package h264

const shortStartCodeLen = 3

// SplitAnnexB attempts to recognize a sequence of NALUs separated by start codes in the buffer.
// Returns a list of raw/unparsed units with emulation bytes removed.
func SplitAnnexB(buf []byte) [][]byte {
	units := [][]byte{}
	end := len(buf) - shortStartCodeLen
	prev := 0
	for off := 0; off < end; {
		if buf[off+2] > 1 {
			// Speed up traversal in search for start codes
			off += 3
		} else if buf[off+2] == 1 && buf[off+1] == 0 && buf[off] == 0 {
			// Do not insert 0-length slices.
			if off > 0 && buf[off-1] == 0 {
				if off > prev+1 {
					units = append(units, EBSPToRaw(buf[prev:off-1]))
				}
			} else if off > prev {
				units = append(units, EBSPToRaw(buf[prev:off]))
			}
			off += 3
			prev = off
		} else {
			off++
		}
	}
	units = append(units, EBSPToRaw(buf[prev:]))
	return units
}

// EBSPToRaw removes emulation prevention bytes from the buffer transforming
// Encapsulated Byte Sequence Payload (EBSP) into Raw Byte Sequence Payload (RBSP).
func EBSPToRaw(buf []byte) []byte {
	zero := byte(0)
	size := len(buf)
	rbsp := make([]byte, 0, size)
	for off := 0; off < size; {
		if size-off >= 3 && buf[off] == 0 && buf[off+1] == 0 && buf[off+2] == 3 {
			// Found emulation prevention byte.
			rbsp = append(rbsp, zero, zero)
			off += 3
		} else {
			rbsp = append(rbsp, buf[off])
			off++
		}
	}
	return rbsp
}

// RBSPToEncapsulated adds emulation prevention bytes to the buffer transforming
// Raw Byte Sequence Payload (RBSP) into Encapsulated Byte Sequence Payload (EBSP).
func RBSPToEncapsulated(buf []byte) []byte {
	zero := byte(0)
	epb := byte(3)
	size := len(buf)
	ebsp := make([]byte, 0, size*2)
	for off := 0; off < size; {
		if size-off >= 2 && buf[off] == 0 && buf[off+1] == 0 && buf[off+2] <= 3 {
			// Add emulation prevention bytes.
			ebsp = append(ebsp, zero, zero, epb)
			off += 2
		}
		ebsp = append(ebsp, buf[off])
		off++
	}
	return ebsp
}
