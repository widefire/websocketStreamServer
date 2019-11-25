package sdp

//https://blog.csdn.net/china_jeffery/article/details/79991986
//rfc 4566
type SessionDescription struct {
	ProtocolVersion int
	Origin          Origin
	SessionName     string
	
}

type Origin struct {
	Username       string
	SessionId      uint64
	SessionVersion uint64
	NetType        string
	AddrType       string
	UnicastAddress string
}
