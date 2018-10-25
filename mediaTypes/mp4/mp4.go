package mp4

const (
	MP4ESDescrTag          = 0x03
	MP4DecConfigDescrTag   = 0x04
	MP4DecSpecificDescrTag = 0x05
)

const (
	CODEC_ID_MOV_TEXT           = 0x08
	CODEC_ID_MPEG4              = 0x20
	CODEC_ID_H264               = 0x21
	CODEC_ID_AAC                = 0x40
	CODEC_ID_MP4ALS             = 0x40 /* 14496-3 ALS */
	CODEC_ID_MPEG2VIDEO_MAIN    = 0x61 /* MPEG2 Main */
	CODEC_ID_MPEG2VIDEO_SIMPLE  = 0x60 /* MPEG2 Simple */
	CODEC_ID_MPEG2VIDEO_SNR     = 0x62 /* MPEG2 SNR */
	CODEC_ID_MPEG2VIDEO_SPATIAL = 0x63 /* MPEG2 Spatial */
	CODEC_ID_MPEG2VIDEO_HIGH    = 0x64 /* MPEG2 High */
	CODEC_ID_MPEG2VIDEO         = 0x65 /* MPEG2 422 */
	CODEC_ID_AAC_MAIN           = 0x66 /* MPEG2 AAC Main */
	CODEC_ID_AAC_LC             = 0x67 /* MPEG2 AAC Low */
	CODEC_ID_AAC_SSR            = 0x68 /* MPEG2 AAC SSR */
	CODEC_ID_MP3_MPEG2          = 0x69 /* 13818-3 */
	CODEC_ID_MP2                = 0x69 /* 11172-3 */
	CODEC_ID_MPEG1VIDEO         = 0x6A /* 11172-2 */
	CODEC_ID_MP3_MPEG1          = 0x6B /* 11172-3 */
	CODEC_ID_MJPEG              = 0x6C /* 10918-1 */
	CODEC_ID_PNG                = 0x6D
	CODEC_ID_JPEG2000           = 0x6E /* 15444-1 */
	CODEC_ID_VC1                = 0xA3
	CODEC_ID_DIRAC              = 0xA4
	CODEC_ID_AC3                = 0xA5
	CODEC_ID_DTS                = 0xA9 /* mp4ra.org */
	CODEC_ID_VORBIS             = 0xDD /* non standard= gpac uses it */
	CODEC_ID_DVD_SUBTITLE       = 0xE0 /* non standard= see unsupported-embedded-subs-2.mp4 */
	CODEC_ID_QCELP              = 0xE1
	CODEC_ID_MPEG4SYSTEMS1      = 0x01
	CODEC_ID_MPEG4SYSTEMS2      = 0x02
	CODEC_ID_NONE               = 0
)
