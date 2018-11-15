package h264

import (
	"math"
)

type (
	// SPSInfo holds Sequence Parameter Set information parsed from out-of-band or in-band data in the stream.
	// Most of the information here is unused but parsed anyway.  VUI parameters are no parsed at all.
	SPSInfo struct {
		ProfileIdc                     byte
		ConstraintSet                  byte
		LevelIdc                       byte
		SpsID                          uint32
		ChromaFormatIdc                uint32
		SeparateColorPlane             bool
		BitDepthLuma                   uint32
		BitDepthChroma                 uint32
		ZeroTransformBypass            bool
		ScalingMatrixPresent           bool
		ScalingListPresent             uint32
		ScalingList                    [6*16 + 6*64]int32
		UseDefaultScalingMatrix        [12]bool
		Log2MaxFrameNum                uint32
		PicOrderCntType                uint32
		Log2MaxPicOrderCnt             uint32
		DeltaPicOrderAlways0           bool
		OffsetForNonRefPic             int32
		OffsetForTopToBottomField      int32
		NumRefFramesInPicOrderCntCycle uint32
		MaxNumRefFrames                uint32
		GapsInFrameNumValueAllowed     bool
		PicWidthInMbsMinus1            uint32
		PicHeightInMapUnitsMinus1      uint32
		FrameMbsOnly                   uint32
		MbAdaptiveFrameField           bool
		Direct8x8Inference             bool
		FrameCropping                  bool
		FrameCropLeftOffset            uint32
		FrameCropRightOffset           uint32
		FrameCropTopOffset             uint32
		FrameCropBottomOffset          uint32
		VuiParametersPresent           bool
		Width                          uint32
		Height                         uint32
	}

	// PPSInfo holds Picture Parameter Set information referenced by slices.
	PPSInfo struct {
		PpsID                              uint32
		SpsID                              uint32
		EntropyCodingMode                  bool
		BottomFieldPicOorderInFramePresent bool
		NumSliceGroupsMinus1               uint32 // 0..7, i.e. up to 8 groups
		SliceGroupMapType                  uint32
		RunLengthMinus1                    [8]uint32
		TopLeft                            [8]uint32
		BottomRight                        [8]uint32
		SliceGroupChangeDirection          bool
		SliceGroupChangeRateMinus1         uint32
		PicSizeInMapUnitsMinus1            uint32
		SliceGroupID                       []byte
		NumRefIdxL0DefaultActiveMinus1     uint32
		NumRefIdxL1DefaultActiveMinus1     uint32
		WeightedPred                       bool
		WeightedBipredIdc                  byte
		PicInitQpMinus26                   int32
		PicInitQsMinus26                   int32
		ChromaQpIndexOffset                int32
		DeblockingFilterControlPresent     bool
		ConstrainedIntraPred               bool
		RedundantPicCntPresent             bool
		Transform8x8Mode                   bool
		ScalingMatrixPresent               bool
		ScalingListPresent                 uint32
		ScalingList                        [6*16 + 6*64]int32
		UseDefaultScalingMatrix            [12]bool
		SecondChromaQpIndexOffset          int32
	}

	// ParameterSets keep all current parsed and indexed parameter sets for quick access.
	ParameterSets struct {
		spset map[uint32]*SPSInfo
		ppset map[uint32]*PPSInfo
	}
)

// NewParameterSets create an representation of indexed storage for sequence
// and picture parameter sets.
func NewParameterSets() *ParameterSets {
	return &ParameterSets{
		spset: make(map[uint32]*SPSInfo),
		ppset: make(map[uint32]*PPSInfo),
	}
}

// ParseSPS parses Sequence Parameter Set information from a given buffer and adds
// it to the indexed list of available sets.
// Buffer can be out-of-band, coming from sprop-parameter-sets property in SDP.
// It could also come in-band in a NAL with respective type.
func (s *ParameterSets) ParseSPS(buf []byte) (err error) {
	sps := &SPSInfo{}
	br := NewBitReader(buf)
	if sps.ProfileIdc, err = br.ReadByteBits(8); err != nil {
		return
	}
	// Leave chroma format at 0 if profile is 183
	if sps.ProfileIdc != 183 {
		sps.ChromaFormatIdc = 1
	}
	if sps.ConstraintSet, err = br.ReadByteBits(8); err != nil {
		return
	}
	// FIXME: Constraint set flags determine certain other values in the absense of overrides.
	if sps.LevelIdc, err = br.ReadByteBits(8); err != nil {
		return
	}
	if sps.SpsID, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if sps.ProfileIdc == 100 || sps.ProfileIdc == 110 || sps.ProfileIdc == 122 || sps.ProfileIdc == 244 ||
		sps.ProfileIdc == 44 || sps.ProfileIdc == 83 || sps.ProfileIdc == 86 || sps.ProfileIdc == 118 ||
		sps.ProfileIdc == 128 || sps.ProfileIdc == 138 {
		if sps.ChromaFormatIdc, err = br.ReadUnsignedGolomb(); err != nil {
			return
		}
		if sps.ChromaFormatIdc == 3 {
			if sps.SeparateColorPlane, err = br.ReadFlag(); err != nil {
				return
			}
		}
		if sps.BitDepthLuma, err = br.ReadUnsignedGolomb(); err != nil {
			return
		}
		if sps.BitDepthChroma, err = br.ReadUnsignedGolomb(); err != nil {
			return
		}
		if sps.ZeroTransformBypass, err = br.ReadFlag(); err != nil {
			return
		}
		if sps.ScalingMatrixPresent, err = br.ReadFlag(); err != nil {
			return
		}
		if sps.ScalingMatrixPresent {
			var present bool
			nlists := 8
			if sps.ChromaFormatIdc == 3 {
				nlists = 12
			}
			for i, off, size := 0, 0, 16; i < nlists; i++ {
				if present, err = br.ReadFlag(); err != nil {
					return
				}
				if present {
					sps.ScalingListPresent |= 1 << uint32(i)
					if sps.UseDefaultScalingMatrix[i], err = br.ReadScalingList(sps.ScalingList[off : off+size]); err != nil {
						return
					}
				}
				off += size
				if i >= 5 {
					size = 64
				}
			}
		}
	}
	if sps.Log2MaxFrameNum, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if sps.PicOrderCntType, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if sps.PicOrderCntType == 0 {
		if sps.Log2MaxPicOrderCnt, err = br.ReadUnsignedGolomb(); err != nil {
			return
		}
	} else {
		if sps.DeltaPicOrderAlways0, err = br.ReadFlag(); err != nil {
			return
		}
		if sps.OffsetForNonRefPic, err = br.ReadSignedGolomb(); err != nil {
			return
		}
		if sps.OffsetForTopToBottomField, err = br.ReadSignedGolomb(); err != nil {
			return
		}
		if sps.NumRefFramesInPicOrderCntCycle, err = br.ReadUnsignedGolomb(); err != nil {
			return
		}
		for i := uint32(0); i < sps.NumRefFramesInPicOrderCntCycle; i++ {
			if _, err = br.ReadSignedGolomb(); err != nil {
				return
			}
		}
	}
	if sps.MaxNumRefFrames, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if sps.GapsInFrameNumValueAllowed, err = br.ReadFlag(); err != nil {
		return
	}
	if sps.PicWidthInMbsMinus1, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if sps.PicHeightInMapUnitsMinus1, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if sps.FrameMbsOnly, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if (sps.ConstraintSet&0x2) != 0 && (sps.ProfileIdc == 77 || sps.ProfileIdc == 88 || sps.ProfileIdc == 100) {
		sps.FrameMbsOnly = 1
	}
	if sps.FrameMbsOnly != 0 {
		if sps.MbAdaptiveFrameField, err = br.ReadFlag(); err != nil {
			return
		}
	}
	if sps.Direct8x8Inference, err = br.ReadFlag(); err != nil {
		return
	}
	if sps.FrameCropping, err = br.ReadFlag(); err != nil {
		return
	}
	if sps.FrameCropping {
		if sps.FrameCropLeftOffset, err = br.ReadUnsignedGolomb(); err != nil {
			return
		}
		if sps.FrameCropRightOffset, err = br.ReadUnsignedGolomb(); err != nil {
			return
		}
		if sps.FrameCropTopOffset, err = br.ReadUnsignedGolomb(); err != nil {
			return
		}
		if sps.FrameCropBottomOffset, err = br.ReadUnsignedGolomb(); err != nil {
			return
		}
	}
	if sps.VuiParametersPresent, err = br.ReadFlag(); err != nil {
		return
	}
	// Macroblock size divisors for chroma as defined in the ITU standard.
	subWidthC := uint32(1)
	subHeightC := uint32(1)
	if !sps.SeparateColorPlane {
		// Chroma format 0 - monochrome, 1 - 4:2:0, 2 - 4:2:2, 3 - 4:4:4
		if sps.ChromaFormatIdc == 1 || sps.ChromaFormatIdc == 2 {
			subWidthC = 2
		}
		if sps.ChromaFormatIdc == 1 {
			subHeightC = 2
		}
	}

	// Macroblock width is 16 for luma.
	sps.Width = (sps.PicWidthInMbsMinus1 + 1) * 16 // / subWidthC
	// FrameMbsOnly designates either full frame or field (half frame).  Hence multiplier.
	sps.Height = (2 - sps.FrameMbsOnly) * (sps.PicHeightInMapUnitsMinus1 + 1) * 16 / subHeightC
	// Adjust for crop if present.  This accounts for monochrome vs. various color chroma formats.
	cropUnitX := subWidthC
	cropUnitY := (2 - sps.FrameMbsOnly) * subHeightC
	sps.Width = sps.Width - (sps.FrameCropLeftOffset+sps.FrameCropRightOffset)*cropUnitX
	sps.Height = sps.Height - (sps.FrameCropTopOffset+sps.FrameCropBottomOffset)*cropUnitY
	s.spset[sps.SpsID] = sps
	return
}

// ParsePPS parses Picture Parameter Set information from a given buffer and adds
// it to the indexed list of available sets.
// Buffer can be out-of-band, coming from sprop-parameter-sets property in SDP.
// It could also come in-band in a NAL with respective type.
func (s *ParameterSets) ParsePPS(buf []byte) (err error) {
	chromaFormatIdc := uint32(3)
	pps := &PPSInfo{}
	br := NewBitReader(buf)
	if pps.PpsID, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if pps.SpsID, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if sps, ok := s.spset[pps.SpsID]; ok {
		chromaFormatIdc = sps.ChromaFormatIdc
	}
	if pps.EntropyCodingMode, err = br.ReadFlag(); err != nil {
		return
	}
	if pps.BottomFieldPicOorderInFramePresent, err = br.ReadFlag(); err != nil {
		return
	}
	if pps.NumSliceGroupsMinus1, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if pps.NumSliceGroupsMinus1 > 0 {
		if pps.SliceGroupMapType, err = br.ReadUnsignedGolomb(); err != nil {
			return
		}
		switch pps.SliceGroupMapType {
		case 0:
			for i := uint32(0); i <= pps.NumSliceGroupsMinus1; i++ {
				if pps.RunLengthMinus1[i], err = br.ReadUnsignedGolomb(); err != nil {
					return
				}
			}
			break
		case 2:
			for i := uint32(0); i < pps.NumSliceGroupsMinus1; i++ {
				if pps.TopLeft[i], err = br.ReadUnsignedGolomb(); err != nil {
					return
				}
				if pps.BottomRight[i], err = br.ReadUnsignedGolomb(); err != nil {
					return
				}
			}
			break
		case 3:
		case 4:
		case 5:
			if pps.SliceGroupChangeDirection, err = br.ReadFlag(); err != nil {
				return
			}
			if pps.SliceGroupChangeRateMinus1, err = br.ReadUnsignedGolomb(); err != nil {
				return
			}
			break
		case 6:
			if pps.PicSizeInMapUnitsMinus1, err = br.ReadUnsignedGolomb(); err != nil {
				return
			}
			pps.SliceGroupID = make([]byte, pps.PicSizeInMapUnitsMinus1+1)
			bits := uint(math.Ceil(math.Log2(float64(pps.NumSliceGroupsMinus1 + 1))))
			for i := uint32(0); i <= pps.PicSizeInMapUnitsMinus1; i++ {
				if pps.SliceGroupID[i], err = br.ReadByteBits(bits); err != nil {
					return
				}
			}
		}
	}
	if pps.NumRefIdxL0DefaultActiveMinus1, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if pps.NumRefIdxL1DefaultActiveMinus1, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if pps.WeightedPred, err = br.ReadFlag(); err != nil {
		return
	}
	if pps.WeightedBipredIdc, err = br.ReadByteBits(2); err != nil {
		return
	}
	if pps.PicInitQpMinus26, err = br.ReadSignedGolomb(); err != nil {
		return
	}
	if pps.PicInitQsMinus26, err = br.ReadSignedGolomb(); err != nil {
		return
	}
	if pps.ChromaQpIndexOffset, err = br.ReadSignedGolomb(); err != nil {
		return
	}
	if pps.DeblockingFilterControlPresent, err = br.ReadFlag(); err != nil {
		return
	}
	if pps.ConstrainedIntraPred, err = br.ReadFlag(); err != nil {
		return
	}
	if pps.RedundantPicCntPresent, err = br.ReadFlag(); err != nil {
		return
	}
	if br.Available() > 0 {
		if pps.Transform8x8Mode, err = br.ReadFlag(); err != nil {
			return
		}
		if pps.ScalingMatrixPresent, err = br.ReadFlag(); err != nil {
			return
		}
		if pps.ScalingMatrixPresent {
			var present bool
			nlists := 6
			if pps.Transform8x8Mode {
				if chromaFormatIdc == 3 {
					nlists = 12
				} else {
					nlists = 8
				}
			}
			for i, off, size := 0, 0, 16; i < nlists; i++ {
				if present, err = br.ReadFlag(); err != nil {
					return
				}
				if present {
					pps.ScalingListPresent |= 1 << uint32(i)
					if pps.UseDefaultScalingMatrix[i], err = br.ReadScalingList(pps.ScalingList[off : off+size]); err != nil {
						return
					}
				}
				off += size
				if i >= 5 {
					size = 64
				}
			}
		}
		if pps.SecondChromaQpIndexOffset, err = br.ReadSignedGolomb(); err != nil {
			return
		}
	}
	s.ppset[pps.PpsID] = pps
	return
}

// ParseSprop analyzes value from SDP sprop parameter sets where first byte is a NAL header.
// To simplify code we do not fully queue and process NAL.
func (s *ParameterSets) ParseSprop(buf []byte) (err error) {
	if buf == nil || len(buf) == 0 {
		return
	}
	typ := buf[0] & 0x1F
	if typ == typeSPS {
		return s.ParseSPS(buf[1:])
	} else if typ == typePPS {
		return s.ParsePPS(buf[1:])
	}
	return
}

// GetSPS looks up SPS by id among currently available parameter sets.
func (s *ParameterSets) GetSPS(id uint32) (sps *SPSInfo, ok bool) {
	sps, ok = s.spset[id]
	return
}

// GetPPS looks up PPS by id among currently available parameter sets.
func (s *ParameterSets) GetPPS(id uint32) (pps *PPSInfo, ok bool) {
	pps, ok = s.ppset[id]
	return
}
