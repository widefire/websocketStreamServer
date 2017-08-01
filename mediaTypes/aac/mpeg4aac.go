package aac

import (
	"fmt"
	"logger"
	"wssAPI"
)

//ADTS : 本地流，需要ADTS header +aac ES
//asc:网络流头,OBJ type就是ADTS的profile+1
//AOT->ASC aac-main=1
//ADTS: aac-main=0
const (
	AOT_NULL = iota
	// Support?                Name
	AOT_AAC_MAIN     ///< Y                       Main
	AOT_AAC_LC       ///< Y                       Low Complexity
	AOT_AAC_SSR      ///< N (code in SoC repo)    Scalable Sample Rate
	AOT_AAC_LTP      ///< Y                       Long Term Prediction
	AOT_SBR          ///< Y                       Spectral Band Replication HE-AAC
	AOT_AAC_SCALABLE ///< N                       Scalable
	AOT_TWINVQ       ///< N                       Twin Vector Quantizer
	AOT_CELP         ///< N                       Code Excited Linear Prediction
	AOT_HVXC         ///< N                       Harmonic Vector eXcitation Coding
)
const (
	AOT_TTSI      = 12 + iota ///< N                       Text-To-Speech Interface
	AOT_MAINSYNTH             ///< N                       Main Synthesis
	AOT_WAVESYNTH             ///< N                       Wavetable Synthesis
	AOT_MIDI                  ///< N                       General MIDI
	AOT_SAFX                  ///< N                       Algorithmic Synthesis and Audio Effects
	AOT_ER_AAC_LC             ///< N                       Error Resilient Low Complexity
)
const (
	AOT_ER_AAC_LTP      = 19 + iota ///< N                       Error Resilient Long Term Prediction
	AOT_ER_AAC_SCALABLE             ///< N                       Error Resilient Scalable
	AOT_ER_TWINVQ                   ///< N                       Error Resilient Twin Vector Quantizer
	AOT_ER_BSAC                     ///< N                       Error Resilient Bit-Sliced Arithmetic Coding
	AOT_ER_AAC_LD                   ///< N                       Error Resilient Low Delay
	AOT_ER_CELP                     ///< N                       Error Resilient Code Excited Linear Prediction
	AOT_ER_HVXC                     ///< N                       Error Resilient Harmonic Vector eXcitation Coding
	AOT_ER_HILN                     ///< N                       Error Resilient Harmonic and Individual Lines plus Noise
	AOT_ER_PARAM                    ///< N                       Error Resilient Parametric
	AOT_SSC                         ///< N                       SinuSoidal Coding
	AOT_PS                          ///< N                       Parametric Stereo
	AOT_SURROUND                    ///< N                       MPEG Surround
	AOT_ESCAPE                      ///< Y                       Escape Value
	AOT_L1                          ///< Y                       Layer 1
	AOT_L2                          ///< Y                       Layer 2
	AOT_L3                          ///< Y                       Layer 3
	AOT_DST                         ///< N                       Direct Stream Transfer
	AOT_ALS                         ///< Y                       Audio LosslesS
	AOT_SLS                         ///< N                       Scalable LosslesS
	AOT_SLS_NON_CORE                ///< N                       Scalable LosslesS (non core)
	AOT_ER_AAC_ELD                  ///< N                       Error Resilient Enhanced Low Delay
	AOT_SMR_SIMPLE                  ///< N                       Symbolic Music Representation Simple
	AOT_SMR_MAIN                    ///< N                       Symbolic Music Representation Main
	AOT_USAC_NOSBR                  ///< N                       Unified Speech and Audio Coding (no SBR)
	AOT_SAOC                        ///< N                       Spatial Audio Object Coding
	AOT_LD_SURROUND                 ///< N                       Low Delay MPEG Surround
	AOT_USAC                        ///< N                       Unified Speech and Audio Coding
)

type MP4AACAudioSpecificConfig struct {
	Object_type        int
	Sampling_index     int
	Sample_rate        int
	Chan_config        int
	Sbr                int ///< -1 implicit, 1 presence int
	Ext_object_type    int
	Ext_sampling_index int
	Ext_sample_rate    int
	Ext_chan_config    int
	Channels           int
	Ps                 int ///< -1 implicit, 1 presence int
	Frame_length_short int
}

func getSampleRatesByIdx(idx int) int {
	arrMpeg4AACSampleRates := [16]int{96000, 88200, 64000, 48000, 44100, 32000,
		24000, 22050, 16000, 12000, 11025, 8000, 7350}
	return arrMpeg4AACSampleRates[idx]
}

func getAudioChannels(idx int) int {
	arr := []int{0, 1, 2, 3, 4, 5, 6, 8}
	return arr[idx]
}

//asc
func MP4AudioGetConfig(data []byte) (asc *MP4AACAudioSpecificConfig) {
	bitReader := &wssAPI.BitReader{}
	bitReader.Init(data)
	asc = &MP4AACAudioSpecificConfig{}
	asc.Object_type = getObjectType(bitReader)
	asc.Sampling_index, asc.Sample_rate = getSampleRate(bitReader)
	asc.Chan_config = bitReader.ReadBits(4)
	if asc.Chan_config < 8 {
		asc.Channels = getAudioChannels(asc.Chan_config)
	}
	asc.Sbr = -1
	asc.Ps = -1
	if AOT_SBR == asc.Object_type || (AOT_PS == asc.Object_type &&
		0 == (bitReader.CopyBits(3)&0x03) && 0 == (bitReader.CopyBits(9)&0x3f)) {
		if AOT_PS == asc.Object_type {
			asc.Ps = 1
		}
		asc.Ext_object_type = AOT_SBR
		asc.Sbr = 1
		asc.Ext_sampling_index, asc.Ext_sample_rate = getSampleRate(bitReader)
		asc.Object_type = getObjectType(bitReader)
		if asc.Object_type == AOT_ER_BSAC {
			asc.Ext_chan_config = bitReader.ReadBits(4)
		}
	} else {
		asc.Ext_object_type = AOT_NULL
		asc.Ext_sample_rate = 0
	}

	if AOT_ALS == asc.Object_type {
		logger.LOGT("ALS")
		bitReader.ReadBits(5)
		als := bitReader.CopyBits(24)
		if ((als>>16)&0xff) != 'A' || ((als>>8)&0xff) != 'L' || ((als)&0xff) != 'S' {
			bitReader.ReadBits(24)
		}
		parseConfigALS(bitReader, asc)

	}

	if asc.Ext_object_type != AOT_SBR {
		//		logger.LOGT(bitReader.BitsLeft())
		for bitReader.BitsLeft() > 15 {
			if 0x2b7 == bitReader.CopyBits(11) {
				bitReader.ReadBits(11)
				asc.Ext_object_type = getObjectType(bitReader)
				if asc.Ext_object_type == AOT_SBR {
					asc.Sbr = bitReader.ReadBit()
					if asc.Sbr == 1 {
						asc.Ext_sampling_index, asc.Ext_sample_rate = getSampleRate(bitReader)
						if asc.Ext_sample_rate == asc.Sample_rate {
							asc.Sbr = -1
						}
					}
					if bitReader.BitsLeft() > 11 && bitReader.ReadBits(11) == 0x548 {
						asc.Ps = bitReader.ReadBit()
					}
					break
				}
			} else {
				bitReader.ReadBit()
			}
		}
	}

	if asc.Sbr == 0 {
		asc.Ps = 0
	}
	if (asc.Ps == -1 && asc.Object_type == AOT_AAC_LC) || (asc.Channels&^0x01) != 0 {
		asc.Ps = 0
	}
	return
}

func CreateAudioSpecificConfigForSDP(asc *MP4AACAudioSpecificConfig) string {
	data := make([]int, 2)

	data[0] = int(byte(asc.Object_type<<3) | byte(asc.Sampling_index>>1))
	data[1] = int(byte(asc.Sampling_index<<7) | byte(asc.Chan_config<<3))
	strConfig := fmt.Sprintf("%02x%02x", data[0], data[1])
	return strConfig
}

func CreateAACADTHeader(asc *MP4AACAudioSpecificConfig, length int) (data []byte) {
	data = make([]byte, 7)
	data[0] = 0xff
	data[1] = 0xf1
	data[2] = byte((((asc.Object_type - 1) & 3) << 6) | (asc.Sampling_index << 2) | (asc.Chan_config >> 2))
	data[3] = byte(((asc.Chan_config & 0x3) << 6) | (((length + 7) >> 11) & 0x3))
	data[4] = (byte((length+7)>>3) & 0xff)
	data[5] = (byte((length+7)&0x7) << 5) | 0x1f
	data[6] = 0xfc
	return
}
func parseConfigALS(bitReader *wssAPI.BitReader, asc *MP4AACAudioSpecificConfig) {
	if bitReader.BitsLeft() < 112 {
		return
	}
	if bitReader.ReadBits(8) != 'A' || bitReader.ReadBits(8) != 'L' || bitReader.ReadBits(8) != 'S' || bitReader.ReadBits(8) != 0 {
		return
	}
	asc.Sample_rate = int(bitReader.Read32Bits())
	bitReader.Read32Bits()
	asc.Chan_config = 0
	asc.Channels = bitReader.ReadBits(16) + 1
}

func getObjectType(bitReader *wssAPI.BitReader) (objectType int) {
	objectType = bitReader.ReadBits(5)
	if objectType == AOT_ESCAPE {
		objectType = 32 + bitReader.ReadBits(6)
	}
	return
}

func getSampleRate(bitReader *wssAPI.BitReader) (sampleFreqIndex, sampleFreq int) {
	sampleFreqIndex = bitReader.ReadBits(4)
	if sampleFreqIndex == 0x0f {
		sampleFreq = bitReader.ReadBits(24)
	} else {
		sampleFreq = getSampleRatesByIdx(sampleFreqIndex)
	}
	return
}
