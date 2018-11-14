package h264

import (
	"testing"
)

func TestParseSPSAmcrestHiDef(t *testing.T) {
	data := []byte{
		0x64, 0x00, 0x1f, 0xac, 0x34, 0xc8, 0x05, 0x00, 0x5b, 0xff, 0x01, 0x6e, 0x02, 0x02, 0x02, 0x80,
		0x00, 0x01, 0xf4, 0x00, 0x00, 0x3a, 0x98, 0x74, 0x30, 0x00, 0x4e, 0x2a, 0x00, 0x01, 0x38, 0xa8,
		0x5d, 0xe5, 0xc6, 0x86, 0x00, 0x09, 0xc5, 0x40, 0x00, 0x27, 0x15, 0x0b, 0xbc, 0xb8, 0x50, 0x00,
	}
	sps, err := parseSPS(data)
	if err != nil {
		t.FailNow()
	}
	// t.Logf("%#v", sps)
	t.Logf("bit_depth_luma=%d, bit_depth_chroma=%d", sps.BitDepthLuma+8, sps.BitDepthChroma+8)
	t.Logf("seq_scaling_list_present=%d, use_default_scaling_matrix=%v", sps.ScalingListPresent, sps.UseDefaultScalingMatrix)
	t.Logf("frame_mbs_only=%d, chroma_format_idc=%d", sps.FrameMbsOnly, sps.ChromaFormatIdc)
	if sps.Width != 1280 {
		t.Errorf("Wrong picture width %d, expected 1280.", sps.Width)
	}
	if sps.Height != 720 {
		t.Errorf("Wrong picture height %d, expected 720.", sps.Height)
	}
}
