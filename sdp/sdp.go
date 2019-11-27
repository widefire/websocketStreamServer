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
	BandWidth          []*BandWidth    //b=* (zero or more bandwidth information lines)
	//One or more time descriptions ("t=" and "r=" lines; see below)
	Timing         Timing         //t= (time the session is active)
	RepeatTimes    []*RepeatTime  //r=* (zero or more repeat times)
	TimeZone       []*TimeZone    //z=* (time zone adjustments)
	EncryptionKeys *EncryptionKey //k=* (encryption key)
	Attributes     []*Attribute   //a=* (zero or more session attribute lines)
	//Zero or more media descriptions
	//Media description, if present
	MediaDescription []*MediaDescription
}

func NewSDP() (sdp *SessionDescription) {
	sdp = &SessionDescription{}
	sdp.BandWidth = make([]*BandWidth, 0)
	sdp.RepeatTimes = make([]*RepeatTime, 0)
	sdp.TimeZone = make([]*TimeZone, 0)
	sdp.Attributes = make([]*Attribute, 0)
	sdp.MediaDescription = make([]*MediaDescription, 0)
	return
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
	if len(values) != 6 {
		err = errors.New("origin must have 6 fields")
		return
	}

	for _, v := range values {
		if len(v) == 0 {
			err = errors.New("origin field can not empty")
			return
		}
	}

	self.Username = values[0]
	self.SessionId, err = strconv.ParseUint(values[1], 10, 64)
	if err != nil {
		return
	}
	self.SessionVersion, err = strconv.ParseUint(values[2], 10, 64)
	if err != nil {
		return
	}
	self.NetType = values[3]
	self.AddrType = values[4]
	self.UnicastAddress = values[5]

	return
}

type ConnectionData struct {
	NetType           string
	AddrType          string
	ConnectionAddress *ConnectionAddress
}

func (self *ConnectionData) Init(line string) (err error) {
	if !strings.HasPrefix(line, "c=") {
		err = errors.New(fmt.Sprintf("invalid connection line %s", line))
		return
	}
	payload := strings.TrimPrefix(line, "c=")
	values := strings.Split(payload, " ")
	if len(values) != 3 {
		err = errors.New("origin must have 3 fields")
		return
	}

	for _, v := range values {
		if len(v) == 0 {
			err = errors.New("connection field can not empty")
			return
		}
	}

	self.NetType = values[0]
	self.AddrType = values[1]
	self.ConnectionAddress = values[2]

	return
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

func InitBandWidthFromLine(line string) (bw *BandWidth, err error) {
	if !strings.HasPrefix(line, "b=") {
		err = errors.New(fmt.Sprintf("invalid band width line %s", line))
		return
	}
	payload := strings.TrimPrefix(line, "b=")
	values := strings.Split(payload, ":")
	if len(values) != 2 || len(values[0]) == 0 || len(values[1]) == 0 {
		err = errors.New("empty band width")
		return
	}
	bw = &BandWidth{}
	bw.BandWidthType = values[0]
	bw.BandWidth, err = strconv.Atoi(values[1])
	return
}

/*
t=3034423619 3042462419
r=604800 3600 0 90000
r=7d 1h 0 25h
*/
type Timing struct {
	StartTime uint64
	StopTime  uint64
}

type RepeatTime struct {
	RepeatInterval       uint64
	ActiveDuration       uint64
	OffsetsFromStartTime []uint64
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
	BandWidth          []*BandWidth    //b=* (zero or more bandwidth information lines)
	EncryptionKeys     *EncryptionKey  //k=* (encryption key)
	Attributes         []*Attribute    //a=* (zero or more session attribute lines)
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

	var currentMediaDesc *MediaDescription
	currentMediaDesc = nil

	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r\n")
		//end
		if len(line) == 0 {
			break
		}
		if len(line) < 3 || !(line[0] >= 'a' && line[0] <= 'z') || line[1] != '=' {

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
			if hasSession {
				err = errors.New("a sdp only one session name")
			} else {
				sdp.SessionName = line[2:]
				hasSession = true
			}
		case 'u':
			if currentMediaDesc != nil {
				err = errors.New("a sdp URI must before media desc")
				return
			}
			if sdp.URI == nil {
				sdp.URI = new(string)
				*sdp.URI = line[2:]
			} else {
				err = errors.New("a sdp can only one URI")
				return
			}
		case 'e':
			if currentMediaDesc != nil {
				err = errors.New("a email must before media desc")
				return
			}
			if sdp.EmailAddress == nil {
				sdp.EmailAddress = new(string)
				*sdp.EmailAddress = line[2:]
			} else {
				err = errors.New("a sdp can only one email")
				return
			}
		case 'p':
			if currentMediaDesc != nil {
				err = errors.New("a phone must before media desc")
				return
			}
			if sdp.PhoneNumber == nil {
				sdp.PhoneNumber = new(string)
				*sdp.PhoneNumber = line[2:]
			} else {
				err = errors.New("a sdp can only one phone")
				return
			}
		case 'z':
		case 't':
			strT := line[2:]
			beginTendT := strings.Split(strT, " ")
			if len(beginTendT) != 2 {
				err = errors.New("")
				return
			}
			sdp.Timing.StartTime, err = strconv.ParseUint(beginTendT[0], 10, 64)
			if err != nil {
				return
			}
			sdp.Timing.StopTime, err = strconv.ParseUint(beginTendT[0], 10, 64)
			if err != nil {
				return
			}
		case 'r':
		case 'm':
		case 'i':
			if currentMediaDesc != nil {
				currentMediaDesc.SessionInformation = new(string)
				*currentMediaDesc.SessionInformation = line[2:]

			} else {

				*sdp.SessionInformation = line[2:]
			}
		case 'c':
			if currentMediaDesc != nil {
				currentMediaDesc.ConnectionData = &ConnectionData{}
				err = currentMediaDesc.ConnectionData.Init(line)
				if err != nil {
					return
				}
			} else {
				sdp.ConnectionData = &ConnectionData{}
				err = sdp.ConnectionData.Init(line)
				if err != nil {
					return
				}
			}
		case 'b':
			if currentMediaDesc != nil {
				bw, err := InitBandWidthFromLine(line)
				if err != nil {
					return nil, err
				}
				currentMediaDesc.BandWidth = append(currentMediaDesc.BandWidth, bw)
			} else {
				bw, err := InitBandWidthFromLine(line)
				if err != nil {
					return nil, err
				}
				sdp.BandWidth = make([]*BandWidth, 0)
				sdp.BandWidth = append(sdp.BandWidth, bw)
			}
		case 'k':
		case 'a':
		}
	}

	if !hasVersion || !hasOirgin || !hasSession {
		return nil, errors.New("invalid sdp")
	}

	return
}
