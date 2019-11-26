package sdp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

//https://blog.csdn.net/china_jeffery/article/details/79991986
//https://tools.ietf.org/id/draft-nandakumar-rtcweb-sdp-01.html
//https://max.book118.com/html/2017/1201/142356242.shtm
//rfc 4566
/*
1 must one
o none or one
*+ none or more
1+ one or more
*/
type SessionDescription struct {
	ProtocolVersion    int             //v= (protocol version)
	Origin             Origin          //o= (originator and session identifier)
	SessionName        string          //s= (session name)
	SessionInformation *string         //i=* (session information)
	URI                *string         //u=* (URI of description)
	EmailAddress       *string         //e=* (email address)
	PhoneNumber        *string         //p=* (phone number)
	ConnectionData     *ConnectionData //c=* (connection information -- not required if included inall media)
	BandWidth          []BandWidth     //b=* (zero or more bandwidth information lines)
	//One or more time descriptions ("t=" and "r=" lines; see below)
	Timing         []Timing       //t= (time the session is active)
	RepeatTimes    []RepeatTime   //r=* (zero or more repeat times)
	TimeZone       []TimeZone     //z=* (time zone adjustments)
	EncryptionKeys *EncryptionKey //k=* (encryption key)
	Attributes     []Attribute    //a=* (zero or more session attribute lines)
	//Zero or more media descriptions
	//Media description, if present
	MediaDescription []MediaDescription
}

type Origin struct {
	Username       string
	SessionId      uint64
	SessionVersion uint64
	NetType        string
	AddrType       string
	UnicastAddress string
}

func (self *Origin) Init(line string) (err error) {
	if !strings.HasPrefix(line, "o=") {
		err = errors.New(fmt.Sprintf("invalid origin line %s", line))
		return
	}
	payload := strings.TrimPrefix(line, "o=")
	values := strings.Split(payload, " ")
	return
}

type ConnectionData struct {
	NetType           string
	AddrType          string
	ConnectionAddress string
}

type ConnectionAddress struct {
	Address           string
	TTL               *int
	NumberOfAddresses *int
}

type BandWidth struct {
	BandWidthType string
	BandWidth     int
}

/*
t=3034423619 3042462419
r=604800 3600 0 90000
r=7d 1h 0 25h
*/
type Timing struct {
	StartTime int64
	StopTime  int64
}

type RepeatTime struct {
	RepeatInterval       int64
	ActiveDuration       int64
	OffsetsFromStartTime []int64
}

type TimeZone struct {
	AdjustmentTime []int64
	Offset         []int64
}

type EncryptionKey struct {
	Method        string
	EncryptionKey string
}

type Attribute struct {
	Attribute string
	Value     string
}

type MediaDescription struct {
	//m= (media name and transport address)
	Media              string
	Port               int
	PortCount          *int
	Protos             []string
	Fmts               []string
	SessionInformation *string         //i=* (media title)
	ConnectionData     *ConnectionData //c=* (connection information -- optional if included at session level)
	BandWidth          []BandWidth     //b=* (zero or more bandwidth information lines)
	EncryptionKeys     *EncryptionKey  //k=* (encryption key)
	Attributes         []Attribute     //a=* (zero or more session attribute lines)
}

func ParseSdp(sdpbuf string) (sdp *SessionDescription, err error) {

	lines := strings.Split(sdpbuf, "\r\n")
	if len(lines) < 3 {
		err = errors.New("a sdp need at least v o s")
		return
	}

	sdp = &SessionDescription{}

	hasVersion := false
	hasOirgin := false
	hasSession := false
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r\n")
		//end
		if len(line) == 0 {
			break
		}
		if len(line) < 3 || !(line[0] >= 'a' && line[0] <= 'z') || line[1] != '=' || line[2] == ' ' {

			err = errors.New(fmt.Sprintf("bad sdp line %s", line))
			return
		}
		lineType := line[0]
		switch lineType {
		case 'v':
			if hasVersion {
				err = errors.New("a sdp only one version")
				return
			} else {
				sdp.ProtocolVersion, err = strconv.Atoi(line[2:])
				if err != nil {
					return
				}
				if sdp.ProtocolVersion != 0 {
					err = errors.New("sdp version only support 0 now")
					return
				}
				hasVersion = true
			}
		case 'o':
			if hasOirgin {
				err = errors.New("a sdp only one Origin")
				return
			} else {
				err = sdp.Origin.Init(line)
				if err != nil {
					return
				}
				hasOirgin = true
			}
		case 's':
		case 'u':
		case 'e':
		case 'p':
		case 'z':
		case 't':
		case 'r':
		case 'm':
		case 'i':
		case 'c':
		case 'b':
		case 'k':
		case 'a':
		}
	}

	if !hasVersion || !hasOirgin || !hasSession {
		return nil, errors.New("invalid sdp")
	}

	return
}
