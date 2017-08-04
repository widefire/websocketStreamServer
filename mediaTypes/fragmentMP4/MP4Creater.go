package fragmentMP4

import "mediaTypes/flv"

func ftypBox()(box *MP4Box)  {
	box=&MP4Box{}
	box.Init([]byte("ftyp"))
	box.PushBytes([]byte("iso5"))
	box.Push4Bytes(1)
	box.PushBytes([]byte("iso5"))
	box.PushBytes([]byte("dash"))
	return box
}

func moovBox(audioTag ,vdeoTag *flv.FlvTag)(box *MP4Box)  {
	box=&MP4Box{}
	box.Init([]byte("moov"))
	//mvhd

	//mvex
	//tracks
	return
}

func mvhdBox()(box *MP4Box)  {
	return 
}

func mvexBox(audioTag ,vdeoTag *flv.FlvTag)(box *MP4Box)  {
	return 
}

func trexBox(tag *flv.FlvTag)(box *MP4Box)  {
return
}

func trakBox(tag *flv.FlvTag)(box *MP4Box)  {
return
}