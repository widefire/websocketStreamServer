package aac

import "bytes"

const (
	AAC_Main      = 1
	AAC_LC        = 2
	AAC_SSR       = 3
	AAC_LTP       = 4
	AAC_HE_OR_SBR = 5
	AAC_SCALABLE  = 6
	ADT_Main      = 0
	ADT_LC        = 1
	ADT_SSR       = 2
)

type AudioSpecificConfig struct {
	AudioObjectType        byte
	SamplingFrequencyIndex byte
	SamplingFrequency      int32
	ChannelConfiguration   byte
	ExtensionSamplingIndex byte
	ExtensionObjectType    byte
}

//从aac头提取audiospecificconfig
func GenerateAudioSpecificConfig(data []byte) (asc AudioSpecificConfig) {
	cur := 0
	asc.AudioObjectType = (data[0] & 0xf8) >> 3
	asc.SamplingFrequencyIndex = ((data[0] & 0x7) << 1) | (data[1] >> 7)
	if 0xf == asc.SamplingFrequencyIndex {

		cur = 4
		//golang byte is uint8

		var fre0, fre1, fre2 int
		fre0 = int(((data[1] & 0x7f) << 1) | (data[2] >> 7))
		fre1 = int(((data[2] & 0x7f) << 1) | (data[3] >> 7))
		fre2 = int(((data[3] & 0x7f) << 1) | (data[4] >> 7))
		asc.SamplingFrequency = int32((fre0 << 16) | (fre1 << 8) | (fre2))

	} else {
		cur = 1
		switch asc.SamplingFrequencyIndex {
		case 0:
			asc.SamplingFrequency = 96000
		case 1:
			asc.SamplingFrequency = 88200
		case 2:
			asc.SamplingFrequency = 64000
		case 3:
			asc.SamplingFrequency = 48000
		case 4:
			asc.SamplingFrequency = 44100
		case 5:
			asc.SamplingFrequency = 32000
		case 6:
			asc.SamplingFrequency = 24000
		case 7:
			asc.SamplingFrequency = 22050
		case 8:
			asc.SamplingFrequency = 16000
		case 9:
			asc.SamplingFrequency = 12000
		case 0xa:
			asc.SamplingFrequency = 11025
		case 0xb:
			asc.SamplingFrequency = 8000
		case 0xc:
			asc.SamplingFrequency = 7350
		default:
			asc.SamplingFrequency = 44100
		}
	}

	asc.ChannelConfiguration = (data[cur] & 0x78) >> 3

	if asc.AudioObjectType == AAC_HE_OR_SBR {
		asc.ExtensionSamplingIndex = ((data[cur] & 0x7 << 1) | (data[cur+1] >> 7))
		cur++
		asc.ExtensionObjectType = ((data[cur] & 0x7c) >> 2)
	}
	return
}

//生成ADT头
func GenerateADTHeader(asc AudioSpecificConfig, length int) (data []byte) {
	data = make([]byte, 7)
	data[0] = 0xff
	data[1] = 0xf1
	data[2] = (((asc.AudioObjectType - 1) & 3) << 6) | (asc.SamplingFrequencyIndex << 2) | (asc.ChannelConfiguration >> 2)
	data[3] = byte(((asc.ChannelConfiguration & 0x3) << 6) | byte(((length+7)>>11)&0x3))
	data[4] = (byte((length+7)>>3) & 0xff)
	data[5] = (byte((length+7)&0x7) << 5) | 0x1f
	data[6] = 0xfc
	return
}

//create aad file,from flv aac
type AACCreater struct {
	writer *bytes.Buffer
	//asc AudioSpecificConfig
	asc       *MP4AACAudioSpecificConfig
	adtsAdded bool
}

func (this *AACCreater) Init(ascData []byte) {
	if this.writer == nil {
		this.writer = new(bytes.Buffer)
	}
	//this.asc=GenerateAudioSpecificConfig(ascData)
	this.asc = MP4AudioGetConfig(ascData)
}

//first frame should asc
func (this *AACCreater) Add(data []byte) {
	this.writer.Write(CreateAACADTHeader(this.asc, len(data)))
	this.writer.Write(data)
}

func (this *AACCreater) Flush() (data []byte) {
	data = this.writer.Bytes()
	this.writer.Reset()
	return
}
