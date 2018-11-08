package h264

import (
	"testing"
)

func TestSplitAnnexB(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0x80, 0x20, 0xda, 0x01, 0x40, 0x16,
		0xe8, 0x06, 0xd0, 0xa1, 0x35, 0x00, 0x00, 0x00, 0x01, 0x68, 0xce, 0x06,
		0xe2, 0x00, 0x00, 0x00, 0x01, 0x65, 0xb8, 0x40, 0xf0, 0x8c, 0x03, 0xf2,
		0x75, 0x67, 0xad, 0x41, 0x64, 0x24, 0x0e, 0xa0, 0xb2, 0x12, 0x1e, 0xf8}
	expected := [][]byte{
		{0x67, 0x42, 0x80, 0x20, 0xda, 0x01, 0x40, 0x16, 0xe8, 0x06, 0xd0, 0xa1, 0x35},
		{0x68, 0xce, 0x06, 0xe2},
		{0x65, 0xb8, 0x40, 0xf0, 0x8c, 0x03, 0xf2, 0x75, 0x67, 0xad, 0x41, 0x64, 0x24, 0x0e, 0xa0, 0xb2, 0x12, 0x1e, 0xf8}}
	nals := splitAnnexB(data)
	if len(nals) != len(expected) {
		t.Fatalf("Split into wrong number (%d) of NALUs", len(nals))
	}
	for i := range nals {
		if len(nals[i]) != len(expected[i]) {
			t.Fatalf("NALU %d is of wrong length (%d)", i, len(nals[i]))
		}
		for j := range nals[i] {
			if nals[i][j] != expected[i][j] {
				t.Fatalf("Byte at position %d in NALU %d is wrong (%x)", j, i, nals[i][j])
			}
		}
	}
}

func TestParseEBSP(t *testing.T) {
	// SPS for a 640x360 camera capture. Contains emulation byte.
	data := []byte{0x7A, 0x00, 0x1E, 0xBC, 0xD9, 0x40, 0xA0, 0x2F,
		0xF8, 0x98, 0x40, 0x00, 0x00, 0x03, 0x01, 0x80,
		0x00, 0x00, 0x56, 0x83, 0xC5, 0x8B, 0x65, 0x80}
	expected := []byte{0x7A, 0x00, 0x1E, 0xBC, 0xD9, 0x40, 0xA0, 0x2F,
		0xF8, 0x98, 0x40, 0x00, 0x00, 0x01, 0x80, 0x00,
		0x00, 0x56, 0x83, 0xC5, 0x8B, 0x65, 0x80}
	buf := parseEBSP(data)
	if len(buf) != len(expected) {
		t.FailNow()
	}
	for i := range buf {
		if buf[i] != expected[i] {
			t.FailNow()
		}
	}
}
