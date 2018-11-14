package h264

// SPSInfo holds Sequence Parameter Set information parsed from out-of-band or in-band data in the stream.
// Most of the information here is unused but parsed anyway.  VUI parameters are no parsed at all.
type SPSInfo struct {
	ProfileIdc                     byte
	ConstraintSet                  byte
	LevelIdc                       byte
	Id                             uint32
	ChromaFormatIdc                uint32
	SeparateColorPlane             bool
	BitDepthLuma                   uint32
	BitDepthChroma                 uint32
	ZeroTransformBypass            bool
	ScalingMatrixPresent           bool
	ScalingListPresent             uint32
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

func parseSPS(buf []byte) (sps *SPSInfo, err error) {
	sps = &SPSInfo{}
	br := NewBitReader(buf)
	if sps.ProfileIdc, err = br.ReadByte(); err != nil {
		return
	}
	// Leave chroma format at 0 if profile is 183
	if sps.ProfileIdc != 183 {
		sps.ChromaFormatIdc = 1
	}
	if sps.ConstraintSet, err = br.ReadByte(); err != nil {
		return
	}
	// FIXME: Constraint set flags determine certain other values in the absense of overrides.
	if sps.LevelIdc, err = br.ReadByte(); err != nil {
		return
	}
	if sps.Id, err = br.ReadExponentialGolomb(); err != nil {
		return
	}
	if sps.ProfileIdc == 100 || sps.ProfileIdc == 110 || sps.ProfileIdc == 122 || sps.ProfileIdc == 244 ||
		sps.ProfileIdc == 44 || sps.ProfileIdc == 83 || sps.ProfileIdc == 86 || sps.ProfileIdc == 118 ||
		sps.ProfileIdc == 128 || sps.ProfileIdc == 138 {
		if sps.ChromaFormatIdc, err = br.ReadExponentialGolomb(); err != nil {
			return
		}
		if sps.ChromaFormatIdc == 3 {
			if sps.SeparateColorPlane, err = br.ReadFlag(); err != nil {
				return
			}
		}
		if sps.BitDepthLuma, err = br.ReadExponentialGolomb(); err != nil {
			return
		}
		if sps.BitDepthChroma, err = br.ReadExponentialGolomb(); err != nil {
			return
		}
		if sps.ZeroTransformBypass, err = br.ReadFlag(); err != nil {
			return
		}
		if sps.ScalingMatrixPresent, err = br.ReadFlag(); err != nil {
			return
		}
		if sps.ScalingMatrixPresent {
			if sps.ChromaFormatIdc == 3 {
				sps.ScalingListPresent, err = br.Read(12)
			} else {
				sps.ScalingListPresent, err = br.Read(8)
			}
			if err != nil {
				return
			}
		}
	}
	if sps.Log2MaxFrameNum, err = br.ReadExponentialGolomb(); err != nil {
		return
	}
	if sps.PicOrderCntType, err = br.ReadExponentialGolomb(); err != nil {
		return
	}
	if sps.PicOrderCntType == 0 {
		if sps.Log2MaxPicOrderCnt, err = br.ReadExponentialGolomb(); err != nil {
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
		if sps.NumRefFramesInPicOrderCntCycle, err = br.ReadExponentialGolomb(); err != nil {
			return
		}
		for i := uint32(0); i < sps.NumRefFramesInPicOrderCntCycle; i++ {
			if _, err = br.ReadSignedGolomb(); err != nil {
				return
			}
		}
	}
	if sps.MaxNumRefFrames, err = br.ReadExponentialGolomb(); err != nil {
		return
	}
	if sps.GapsInFrameNumValueAllowed, err = br.ReadFlag(); err != nil {
		return
	}
	if sps.PicWidthInMbsMinus1, err = br.ReadExponentialGolomb(); err != nil {
		return
	}
	if sps.PicHeightInMapUnitsMinus1, err = br.ReadExponentialGolomb(); err != nil {
		return
	}
	if sps.FrameMbsOnly, err = br.ReadExponentialGolomb(); err != nil {
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
		if sps.FrameCropLeftOffset, err = br.ReadExponentialGolomb(); err != nil {
			return
		}
		if sps.FrameCropRightOffset, err = br.ReadExponentialGolomb(); err != nil {
			return
		}
		if sps.FrameCropTopOffset, err = br.ReadExponentialGolomb(); err != nil {
			return
		}
		if sps.FrameCropBottomOffset, err = br.ReadExponentialGolomb(); err != nil {
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
	return sps, nil
}
