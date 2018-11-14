package h264

import (
	"math"
)

type (
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
)

// ParsePPS retrieves Picture Parameter Set information from a given buffer.
// Buffer can be out-of-band, coming from sprop-parameter-sets property in SDP.
// Ic ould also come in-band in a NAL with respective type.
func ParsePPS(buf []byte, chromaFormatIdc uint32) (pps *PPSInfo, err error) {
	pps = &PPSInfo{}
	br := NewBitReader(buf)
	if pps.PpsID, err = br.ReadUnsignedGolomb(); err != nil {
		return
	}
	if pps.SpsID, err = br.ReadUnsignedGolomb(); err != nil {
		return
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
	return
}
