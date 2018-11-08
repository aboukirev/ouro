package h264

const shortStartCodeLen = 3

// splitAnnexB attempts to recognize a sequence of NALUs separated by start codes in the buffer.
// Returns a list of raw/unparsed units with emulation bytes removed.
func splitAnnexB(buf []byte) [][]byte {
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
					units = append(units, parseEBSP(buf[prev:off-1]))
				}
			} else if off > prev {
				units = append(units, parseEBSP(buf[prev:off]))
			}
			off += 3
			prev = off
		} else {
			off++
		}
	}
	units = append(units, parseEBSP(buf[prev:]))
	return units
}

// parseEBSP removes emulation prevention bytes from the buffer transforming
// Encapsulated Byte Sequence Payload (EBSP) into Raw Byte Sequence Payload (RBSP).
func parseEBSP(buf []byte) []byte {
	zero := byte(0)
	size := len(buf)
	// Modified value buffer overlays the original buffer.  No allocation occurs.
	// This works because RBSP is always same or shorter than EBSP.
	rbsp := buf[:0]
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
