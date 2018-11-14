package h264

import (
	"testing"
)

func TestParseSPSHiDef(t *testing.T) {
	data := []byte{0x7A, 0x00, 0x1F, 0xBC, 0xD9, 0x40, 0x50, 0x05,
		0xBA, 0x10, 0x00, 0x00, 0x03, 0x00, 0xC0, 0x00,
		0x00, 0x2A, 0xE0, 0xF1, 0x83, 0x19, 0x60}
	sps, err := parseSPS(data)
	if err != nil {
		t.FailNow()
	}
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

func TestParseSPSStdDef(t *testing.T) {
	data := []byte{0x7A, 0x00, 0x1E, 0xBC, 0xD9, 0x40, 0xA0, 0x2F,
		0xF8, 0x98, 0x40, 0x00, 0x00, 0x03, 0x01, 0x80,
		0x00, 0x00, 0x56, 0x83, 0xC5, 0x8B, 0x65, 0x80}
	sps, err := parseSPS(data)
	if err != nil {
		t.FailNow()
	}
	// t.Logf("%#v", sps)
	t.Logf("bit_depth_luma=%d, bit_depth_chroma=%d", sps.BitDepthLuma+8, sps.BitDepthChroma+8)
	t.Logf("seq_scaling_list_present=%d, use_default_scaling_matrix=%v", sps.ScalingListPresent, sps.UseDefaultScalingMatrix)
	t.Logf("frame_mbs_only=%d, chroma_format_idc=%d", sps.FrameMbsOnly, sps.ChromaFormatIdc)
	if sps.Width != 640 {
		t.Errorf("Wrong picture width %d, expected 640.", sps.Width)
	}
	if sps.Height != 360 {
		t.Errorf("Wrong picture height %d, expected 360.", sps.Height)
	}
}

func TestParseSPSHalfDef(t *testing.T) {
	data := []byte{0x7A, 0x00, 0x0D, 0xBC, 0xD9, 0x43, 0x43, 0x3E,
		0x5E, 0x10, 0x00, 0x00, 0x03, 0x00, 0x60, 0x00,
		0x00, 0x15, 0xA0, 0xF1, 0x42, 0x99, 0x60}
	sps, err := parseSPS(data)
	if err != nil {
		t.FailNow()
	}
	t.Logf("bit_depth_luma=%d, bit_depth_chroma=%d", sps.BitDepthLuma+8, sps.BitDepthChroma+8)
	t.Logf("seq_scaling_list_present=%d, use_default_scaling_matrix=%v", sps.ScalingListPresent, sps.UseDefaultScalingMatrix)
	t.Logf("frame_mbs_only=%d, chroma_format_idc=%d", sps.FrameMbsOnly, sps.ChromaFormatIdc)
	if sps.Width != 200 {
		t.Errorf("Wrong picture width %d, expected 200.", sps.Width)
	}
	if sps.Height != 400 {
		t.Errorf("Wrong picture height %d, expected 400.", sps.Height)
	}
}
