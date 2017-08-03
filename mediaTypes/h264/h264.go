package h264

import (
	"wssAPI"
)

const (
	Nal_type_slice    = 1
	Nal_type_dpa      = 2
	Nal_type_dpb      = 3
	Nal_type_dpc      = 4
	Nal_type_idr      = 5
	Nal_type_sei      = 6
	Nal_type_sps      = 7
	Nal_type_pps      = 8
	Nal_type_aud      = 9
	Nal_type_eoseq    = 10
	Nal_type_eostream = 11
	Nal_type_fill     = 12
)

func ParseSPS(sps []byte) (width, height, fps int) {
	tmpSps := make([]byte, len(sps))
	copy(tmpSps, sps)
	realSPS := EmulationPrevention(sps)

	bit := &wssAPI.BitReader{}
	bit.Init(realSPS)
	bit.ReadBits(8)
	profile_idc := bit.ReadBits(8)
	bit.ReadBits(16)
	bit.ReadExponentialGolombCode()

	if profile_idc == 100 || profile_idc == 110 ||
		profile_idc == 122 || profile_idc == 244 ||
		profile_idc == 44 || profile_idc == 83 ||
		profile_idc == 86 || profile_idc == 118 {
		chroma_format_idc := bit.ReadExponentialGolombCode()
		if chroma_format_idc == 3 {
			bit.ReadBit()
		}
		bit.ReadExponentialGolombCode()//bit_depth_luma_minus
		bit.ReadExponentialGolombCode()//bit_depth_chroma_minus8
		bit.ReadBit()
		seq_scaling_matrix_present_flag := bit.ReadBit()
		if seq_scaling_matrix_present_flag != 0 {
			for i := 0; i < 8; i++ {
				seq_scaling_list_present_flag := bit.ReadBit()
				if seq_scaling_list_present_flag != 0 {
					var sizeOfScalingList int
					if i < 6 {
						sizeOfScalingList = 16
					} else {
						sizeOfScalingList = 64
					}
					lastScale := 8
					nextScale := 8
					for j := 0; j < sizeOfScalingList; j++ {
						delta_scale := bit.ReadSE()
						nextScale = (lastScale + delta_scale + 256) % 256
					}
					if nextScale == 0 {
						lastScale = lastScale
					} else {
						lastScale = nextScale
					}
				}
			}
		}
	}

	bit.ReadExponentialGolombCode()
	pic_order_cnt_type := bit.ReadExponentialGolombCode()
	if 0 == pic_order_cnt_type {
		bit.ReadExponentialGolombCode()
	} else if 1 == pic_order_cnt_type {
		bit.ReadBit()
		bit.ReadSE()
		bit.ReadSE()
		num_ref_frames_in_pic_order_cnt_cycle := bit.ReadExponentialGolombCode()
		for i := 0; i < num_ref_frames_in_pic_order_cnt_cycle; i++ {
			bit.ReadSE()
		}
	}

	bit.ReadExponentialGolombCode()
	bit.ReadBit()
	pic_width_in_mbs_minus1 := bit.ReadExponentialGolombCode()
	pic_height_in_map_units_minus1 := bit.ReadExponentialGolombCode()
	frame_mbs_only_flag := bit.ReadBit()
	if frame_mbs_only_flag == 0 {
		bit.ReadBit()
	}
	bit.ReadBit()
	frame_cropping_flag := bit.ReadBit()
	var frame_crop_left_offset int
	var frame_crop_right_offset int
	var frame_crop_top_offset int
	var frame_crop_bottom_offset int
	if frame_cropping_flag != 0 {
		frame_crop_left_offset = bit.ReadExponentialGolombCode()
		frame_crop_right_offset = bit.ReadExponentialGolombCode()
		frame_crop_top_offset = bit.ReadExponentialGolombCode()
		frame_crop_bottom_offset = bit.ReadExponentialGolombCode()
	}

	width = ((pic_width_in_mbs_minus1 + 1) * 16) - frame_crop_bottom_offset*2 - frame_crop_top_offset*2
	height = ((2 - frame_mbs_only_flag) * (pic_height_in_map_units_minus1 + 1) * 16) - (frame_crop_right_offset * 2) - (frame_crop_left_offset * 2)

	vui_parameters_present_flag := bit.ReadBit()
	if vui_parameters_present_flag != 0 {
		aspect_ratio_info_present_flag := bit.ReadBit()
		if aspect_ratio_info_present_flag != 0 {
			aspect_ratio_idc := bit.ReadBits(8)
			if aspect_ratio_idc == 255 {
				bit.ReadBits(16)
				bit.ReadBits(16)
			}
		}
		overscan_info_present_flag := bit.ReadBit()
		if 0 != overscan_info_present_flag {
			bit.ReadBit()
		}
		video_signal_type_present_flag := bit.ReadBit()
		if video_signal_type_present_flag != 0 {
			bit.ReadBits(3)
			bit.ReadBit()
			colour_description_present_flag := bit.ReadBit()
			if colour_description_present_flag != 0 {
				bit.ReadBits(8)
				bit.ReadBits(8)
				bit.ReadBits(8)
			}
		}
		chroma_loc_info_present_flag := bit.ReadBit()
		if chroma_loc_info_present_flag != 0 {
			bit.ReadExponentialGolombCode()
			bit.ReadExponentialGolombCode()
		}

		timing_info_present_flag := bit.ReadBit()
		if 0 != timing_info_present_flag {
			num_units_in_tick := bit.ReadBits(32)
			time_scale := bit.ReadBits(32)
			fps = time_scale / (2 * num_units_in_tick)
		}
	}
	return
}

func EmulationPrevention(src []byte) (ret []byte) {
	size := len(src)
	tmpLen := size
	for i := 0; i < tmpLen-2; i++ {
		if (src[i] == 0) && (src[i+1] == 0) && (src[i+2] == 3) {
			for j := i + 2; j < tmpLen-1; j++ {
				src[j] = src[j+1]
			}
			size--
		}
	}
	ret = make([]byte, size)
	copy(ret, src)
	return ret
}

//default only one sps pps
func GetSpsPpsFromAVC(avc []byte) (sps, pps []byte) {
	spsSize := ((int(avc[6]) << 8) | (int(avc[7])))
	sps = make([]byte, spsSize)
	copy(sps, avc[8:8+spsSize])
	cur := 8 + spsSize
	cur++
	ppsSize := ((int(avc[cur]) << 8) | (int(avc[cur+1])))
	cur += 2
	pps = make([]byte, ppsSize)
	copy(pps, avc[cur:cur+ppsSize])
	return

}
