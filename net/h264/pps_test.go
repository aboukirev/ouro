package h264

import (
	"testing"
)

func TestParsePPS(t *testing.T) {
	data := []byte{
		0xee, 0x3c, 0x30, 0x00,
	}
	pps, err := parsePPS(data, 1)
	if err != nil {
		t.Error(err)
	}
	// t.Logf("%#v", pps)
	if pps.PpsID != 0 {
		t.Errorf("PPS Id is %d, expected 0", pps.PpsID)
	}
}
