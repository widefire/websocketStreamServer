module github.com/widefire/websocketStreamServer/rtsp

go 1.12

require github.com/writefire/websocketStreamServer/sdp v0.0.0
require github.com/writefire/websocketStreamServer/rtp v0.0.0
require github.com/writefire/websocketStreamServer/rtcp v0.0.0

replace github.com/writefire/websocketStreamServer/sdp => ../sdp
replace github.com/writefire/websocketStreamServer/rtp => ../rtp
replace github.com/writefire/websocketStreamServer/rtcp => ../rtcp