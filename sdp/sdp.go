package sdp

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

//https://blog.csdn.net/china_jeffery/article/details/79991986
//https://tools.ietf.org/id/draft-nandakumar-rtcweb-sdp-01.html
//rfc 4566

//Session description
const (
	SessionDescriptionV rune = 'v' //protocol version
	SessionDescriptionO rune = 'o' //originator and session identifier
	SessionDescriptionS rune = 's' //session name
	SessionDescriptionI rune = 'i' //*session information
	SessionDescriptionU rune = 'u' //*URI of description
	SessionDescriptionE rune = 'e' //*email address
	SessionDescriptionP rune = 'p' //*phone number
	SessionDescriptionC rune = 'c' //*connection information -- not required if included in all media
	SessionDescriptionB rune = 'b' //*zero or more bandwidth information lines
	SessionDescriptionZ rune = 'z' //*time zone adjustments
	SessionDescriptionK rune = 'k' //*encryption key
	SessionDescriptionA rune = 'a' //*zero or more session attribute lines
	SessionDescriptionT rune = 't' //time the session is active
	SessionDescriptionR rune = 'r' //*zero or more repeat times
	SessionDescriptionM rune = 'm' //*media name and transport address
)

//SessionDescription ...
type SessionDescription struct {
	ProtocolVersion    *int            //v= (protocol version)
	Origin             *Origin         //o= (originator and session identifier)
	SessionName        *string         //s= (session name)
	SessionInformation *string         //i=* (session information)
	URI                *string         //u=* (URI of description)
	EmailAddress       []string        //e=* (email address)
	PhoneNumber        []string        //p=* (phone number)
	ConnectionData     *ConnectionData //c=* (connection information -- not required if included inall media)
	Bandwidths         []*Bandwidth    //b=* (zero or more bandwidth information lines)
	//One or more time descriptions ("t=" and "r=" lines; see below)
	Timing            *Timing        //t= (time the session is active)
	RepeatTimes       []*RepeatTime  //r=* (zero or more repeat times)
	TimeZones         []*TimeZone    //z=* (time zone adjustments)
	EncryptionKey     *EncryptionKey //k=* (encryption key)
	Attributes        []*Attribute   //a=* (zero or more session attribute lines)
	MediaDescriptions []*MediaLevel
}

//NewSDP ...
func NewSDP() *SessionDescription {
	return &SessionDescription{
		ProtocolVersion:    nil,
		Origin:             nil,
		SessionName:        nil,
		SessionInformation: nil,
		URI:                nil,
		EmailAddress:       make([]string, 0),
		PhoneNumber:        make([]string, 0),
		ConnectionData:     nil,
		Bandwidths:         make([]*Bandwidth, 0),
		Timing:             nil,
		RepeatTimes:        make([]*RepeatTime, 0),
		TimeZones:          make([]*TimeZone, 0),
		EncryptionKey:      nil,
		Attributes:         make([]*Attribute, 0),
		MediaDescriptions:  make([]*MediaLevel, 0),
	}
}

/*
Version ..
5.1. Protocol Version ("v=")
v=0
one
*/
type Version int32

/*
Origin ..
5.2. Origin ("o=")
o=<username> <sess-id> <sess-version> <nettype> <addrtype>
<unicast-address>
one
*/
type Origin struct {
	Username       string //"-" for empty username
	SessionID      string //numeric string
	SessionVersion string //numeric string
	Nettype        string //"IN" or some other
	Addrtype       string //"IP4" or "IP6"
	UnicastAddress string //gen session domain name or IP4/6 addr,don't use local IP
}

//NettypeIn ..."Internet
const NettypeIn = "IN"

/*
SessionName ...
5.3. Session Name ("s=")
s=<session name>
one
*/
type SessionName string

/*
SessionInformation ...
5.4. Session Information ("i=")
i=<session description>
at most one session-level "i=" field per session description,
and at most one "i=" field per media
*/
type SessionInformation string

/*
URI ...
5.5. URI ("u=")
u=<uri>
at most one , before media
*/
type URI string

/*
EmailAddress ...
5.6. Email Address and Phone Number ("e=" and "p=")
zero or more , before media
*/
type EmailAddress string

/*
PhoneNumber ...
5.6. Email Address and Phone Number ("e=" and "p=")
zero or more ,  before media
*/
type PhoneNumber string

/*
ConnectionAddressDesc ...
multicast
	IP4	addr/ttl/(number)
	IP6	addr/(number)
unicast
	IP4 	addr
	IP6	addr
multicast addr can't in session level
*/
type ConnectionAddressDesc struct {
	Addr            string
	TTL             *int
	NumberOfAddress *int
}

/*
ConnectionData ...
5.7. Connection Data ("c=")
c=<nettype> <addrtype> <connection-address>
each media at least one or session level must one
*/
type ConnectionData struct {
	Nettype           string //IN
	Addrtype          string //IP4 or IP6
	ConnectionAddress string
}

/*
Bandwidth ...
5.8. Bandwidth ("b=")
b=<bwtype>:<bandwidth>
bwtype:
	CT : total bandwidth
	AS : one RTP bandwidth
	X- : experimental purposes
default kilobits per second
optional , zero or more
*/
type Bandwidth struct {
	Bwtype         string
	BandwidthValue uint64
}

/*
Timing ...
5.9. Timing ("t=")
t=<start-time> <stop-time>
NPT =UNIX time + 2208988800 seconds
if StopTime=0 , after startTime not bounded
if StartTime=0 , permanent
if permanent ,assumption a user half an hour before active
*/
type Timing struct {
	StartTime uint64
	StopTime  uint64
}

/*
RepeatTime ...
5.10. Repeat Times ("r=")
r=<repeat interval> <active duration> <offsets from start-time>
d h m s
default s
*/
type RepeatTime struct {
	RepeatInterval       string
	ActiveDuration       string
	OffsetsFromStartTime []string
}

//NewRepeatTime ...
func NewRepeatTime() *RepeatTime {
	return &RepeatTime{
		RepeatInterval:       "",
		ActiveDuration:       "",
		OffsetsFromStartTime: make([]string, 0),
	}
}

/*
TimeZone ...
5.11. Time Zones ("z=")
夏令时
z=<adjustment time> <offset> <adjustment time> <offset> ....
在某个时刻，调整基准时间
*/
type TimeZone struct {
	AdjustmentTime []uint64
	Offset         []string
}

//NewTimeZone ...
func NewTimeZone() *TimeZone {
	return &TimeZone{
		AdjustmentTime: make([]uint64, 0),
		Offset:         make([]string, 0),
	}
}

/*
EncryptionKey ...
5.12. Encryption Keys ("k=")
k=<method>
k=<method>:<encryption key>
sdp level for all media,media level for media
optional, NOT RECOMMENDED
*/
type EncryptionKey struct {
	Method        string
	EncryptionKey *string
}

//EncryptionMethodClear ... clear
const EncryptionMethodClear = "clear"

//EncryptionMethodBase64 ... base64
const EncryptionMethodBase64 = "base64"

//EncryptionMethodURI ... get key from uri
const EncryptionMethodURI = "uri"

//EncryptionMethodPrompt ... prompt
const EncryptionMethodPrompt = "prompt"

/*
Attribute ...
5.13. Attributes ("a=")
a=<attribute>	flag
a=<attribute>:<value> k:v
*/
type Attribute struct {
	AttributeName string
	Value         *string
}

/*
MediaDescription ...
5.14. Media Descriptions ("m=")
m=<media> <port> <proto> <fmt> ...
*/
type MediaDescription struct {
	Media         string //media type:"audio","video", "text", "application", and "message"
	Port          int    //defaut RTP odd,RTCP=RTP+1
	NumberOfPorts *int
	Proto         string
	Fmt           []string
}

//NewMediaDescription ...
// func NewMediaDescription() *MediaDescription {
// 	instance := &MediaDescription{}
// 	instance.Fmt = make([]string, 0)
// 	return instance
// }

//MediaProtoUDP ... udp
const MediaProtoUDP = "udp"

//MediaProtoRTPAVP ... RTP/AVP
const MediaProtoRTPAVP = "RTP/AVP"

//MediaProtoRTPSAVP ... RTP/SAVP
const MediaProtoRTPSAVP = "RTP/SAVP"

//MediaLevel ...
type MediaLevel struct {
	MediaDescription
	SessionInformation *string
	ConnectionData     *ConnectionData
	BandWidths         []*Bandwidth
	EncryptionKey      *EncryptionKey
	Attribute          []*Attribute
}

//NewMediaLevel ...
func NewMediaLevel() *MediaLevel {
	mediaLevel := &MediaLevel{}
	mediaLevel.Fmt = make([]string, 0)
	mediaLevel.BandWidths = make([]*Bandwidth, 0)
	mediaLevel.Attribute = make([]*Attribute, 0)
	return mediaLevel
}

//Attributes key
const (
	AttrCat       = "cat"      //category
	AttrKeywds    = "keywds"   //keywords
	AttrTool      = "tool"     //name and version of tool
	AttrPtime     = "ptime"    //audio pkt length in mill
	AttrMaxptime  = "maxptime" //max packet time in mill
	AttrRtpmap    = "rtpmap"   //<payload type> <encoding name>/<clock rate> [/<encoding parameters>]
	AttrRecvonly  = "recvonly"
	AttrSendRecv  = "sendrecv"
	AttrSendOnly  = "sendoly"
	AttrInactive  = "inactive"
	AttrType      = "type" //<conference type>Suggested values are "broadcast", "meeting", "moderated", "test", and "H332"
	AttrCharset   = "charset"
	AttrSdplang   = "sdplang"
	AttrLang      = "lang"
	AttrFramerate = "framerate" //max video frame /sec
	AttrQuality   = "quality"   //0-10  10is best
	AttrFmtp      = "fmtp"
)

//RTPMAP ...
type RTPMAP struct {
	PayloadType        string
	EncodingName       string
	ClockRate          uint32
	EncodingParameters *string
}

//ParseSDP ... parse a sdp string
func ParseSDP(sdpPayload string) (sdp *SessionDescription, err error) {
	lines := strings.Split(sdpPayload, "\r\n")
	if len(lines) < 3 {
		err = errors.New("a sdp need at least v o s")
		return
	}
	sdp = NewSDP()

	var curMediaLevel *MediaLevel
	for index, line := range lines {
		line = strings.TrimSuffix(line, "\r\n")
		if len(line) == 0 {
			break
		}
		if len(line) < 3 || !(line[0] >= 'a' && line[0] <= 'z') || line[1] != '=' {
			err = fmt.Errorf("invalid sdp line %s", line)
			continue
		}
		if index == 0 {
			if line[0] == 'v' {
				version, err := strconv.Atoi(line[2:])
				if err != nil {
					return nil, err
				}
				sdp.ProtocolVersion = new(int)
				*sdp.ProtocolVersion = version
			} else {
				return nil, errors.New("first line not v=")
			}

		} else {
			key := line[0]
			switch key {
			case 'o':
				if sdp.Origin == nil {
					values := strings.Split(line[2:], " ")
					if len(values) != 6 {
						return nil, fmt.Errorf("invalid o= %s", line[2:])
					}
					sdp.Origin = &Origin{
						Username:       values[0],
						SessionID:      values[1],
						SessionVersion: values[2],
						Nettype:        values[3],
						Addrtype:       values[4],
						UnicastAddress: values[5],
					}
				} else {
					return nil, errors.New("a sdp at most one origin o=")
				}
			case 's':
				if sdp.SessionName == nil {
					sdp.SessionName = new(string)
					*sdp.SessionName = line[2:]
				} else {
					return nil, errors.New("a sdp at most one session name s=")
				}
			case 'u':
				if curMediaLevel == nil {
					if sdp.URI == nil {
						sdp.URI = new(string)
						*sdp.URI = line[2:]
					} else {
						return nil, errors.New("a sdp at most one URI")
					}
				} else {
					return nil, errors.New("URI must before media")
				}
			case 'e':
				if curMediaLevel == nil {
					sdp.EmailAddress = append(sdp.EmailAddress, line[2:])
				} else {
					return nil, errors.New("email must before media")
				}
			case 'p':
				if curMediaLevel == nil {
					sdp.PhoneNumber = append(sdp.PhoneNumber, line[2:])
				} else {
					return nil, errors.New("email must before media")
				}
			case 't':
				if sdp.Timing == nil {
					values := strings.Split(line[2:], " ")
					if len(values) != 2 {
						log.Println(line)
						continue
					}
					timing := &Timing{}
					timing.StartTime, err = strconv.ParseUint(values[0], 10, 64)
					if err != nil {
						log.Println(err)
						log.Println(line)
						continue
					}
					timing.StopTime, err = strconv.ParseUint(values[1], 10, 64)
					if err != nil {
						log.Println(err)
						log.Println(line)
						continue
					}
					sdp.Timing = timing
				} else {
					return nil, errors.New("a sdp must one timing")
				}
			case 'r':
				values := strings.Split(line[2:], " ")
				if len(values) < 2 {
					log.Println(line)
					continue
				}
				repeatTime := NewRepeatTime()
				repeatTime.RepeatInterval = values[0]
				repeatTime.ActiveDuration = values[1]
				for i := 2; i < len(values); i++ {
					repeatTime.OffsetsFromStartTime = append(repeatTime.OffsetsFromStartTime, values[1])
				}
				sdp.RepeatTimes = append(sdp.RepeatTimes, repeatTime)
			case 'z':
				values := strings.Split(line[2:], " ")
				if len(values) == 0 {
					log.Println(line)
					continue
				}
				if len(values)%2 != 0 {
					log.Println(line)
					continue
				}
				timezone := NewTimeZone()
				for i := 0; i < len(values); i += 2 {
					adjustment, err := strconv.ParseUint(values[i], 10, 64)
					if err != nil {
						log.Println(err)
						log.Println(line)
						continue
					}
					timezone.AdjustmentTime = append(timezone.AdjustmentTime, adjustment)
					timezone.Offset = append(timezone.Offset, values[i+1])
				}
				sdp.TimeZones = append(sdp.TimeZones, timezone)
			case 'm':
				values := strings.Split(line[2:], " ")
				if len(values) < 4 {
					log.Println(line)
					continue
				}
				mediaLevel := NewMediaLevel()
				mediaLevel.Media = values[0]
				pp := strings.Split(values[1], "/")
				if len(pp) == 1 {
					mediaLevel.Port, err = strconv.Atoi(pp[0])
					if err != nil {
						log.Println(err)
						log.Println(line)
						continue
					}
				} else if len(pp) == 2 {
					mediaLevel.Port, err = strconv.Atoi(pp[0])
					if err != nil {
						log.Println(err)
						log.Println(line)
						continue
					}
					mediaLevel.NumberOfPorts = new(int)
					*mediaLevel.NumberOfPorts, err = strconv.Atoi(pp[1])
					if err != nil {
						log.Println(err)
						log.Println(line)
						continue
					}
				} else {
					log.Println(line)
					continue
				}

				mediaLevel.Proto = values[2]
				slice := values[3:]
				mediaLevel.Fmt = append(mediaLevel.Fmt, slice...)
				sdp.MediaDescriptions = append(sdp.MediaDescriptions, mediaLevel)
				curMediaLevel = mediaLevel
				//next keys both session level and media level use
			case 'i':
				if curMediaLevel != nil {
					if curMediaLevel.SessionInformation == nil {
						curMediaLevel.SessionInformation = new(string)
						*curMediaLevel.SessionInformation = line[2:]
					} else {
						return nil, errors.New(" a media at most one per media")
					}
				} else {
					if sdp.SessionInformation == nil {
						sdp.SessionInformation = new(string)
						*sdp.SessionInformation = line[2:]
					} else {
						return nil, errors.New("a sdp level at most one per session")
					}
				}
			case 'c':
				values := strings.Split(line[2:], " ")
				if len(values) < 3 {
					log.Println(line)
					continue
				}
				connectionData := &ConnectionData{
					Nettype:           values[0],
					Addrtype:          values[1],
					ConnectionAddress: values[2],
				}
				for i := 3; i < len(values); i++ {
					connectionData.ConnectionAddress += values[i]
				}

				if curMediaLevel != nil {
					if curMediaLevel.ConnectionData == nil {
						curMediaLevel.ConnectionData = connectionData
					} else {
						return nil, errors.New("a media connection data at most one")
					}
				} else {
					if sdp.ConnectionData == nil {
						sdp.ConnectionData = connectionData
					} else {
						return nil, errors.New("a sdp connection data at most one")
					}
				}
			case 'b':
				values := strings.Split(line[2:], ":")
				if len(values) != 2 {
					log.Println(line)
					continue
				}
				bandwidth := &Bandwidth{}

				bandwidth.Bwtype = values[0]
				bandwidth.BandwidthValue, err = strconv.ParseUint(values[1], 10, 64)
				if err != nil {
					log.Println(line)
					continue
				}
				if curMediaLevel != nil {
					curMediaLevel.BandWidths = append(curMediaLevel.BandWidths, bandwidth)
				} else {
					sdp.Bandwidths = append(sdp.Bandwidths, bandwidth)
				}
			case 'k':
				values := strings.Split(line[2:], ":")
				encryptionKey := &EncryptionKey{}

				if len(values) == 0 {
					log.Println(line)
					continue
				}
				encryptionKey.Method = values[0]
				if len(values) > 1 {
					encryptionKey.EncryptionKey = new(string)
					var key string
					for i := 1; i < len(values); i++ {
						key += values[i]
					}
					*encryptionKey.EncryptionKey = key
				}

				if curMediaLevel != nil {
					if curMediaLevel.EncryptionKey == nil {
						curMediaLevel.EncryptionKey = encryptionKey
					} else {
						return nil, errors.New("a media at most one encryption key")
					}
				} else {
					if sdp.EncryptionKey == nil {
						sdp.EncryptionKey = encryptionKey
					} else {
						return nil, errors.New("a sdp at most one encryption key")
					}
				}
			case 'a':
				values := strings.Split(line[2:], ":")
				if len(values) == 0 {
					log.Println(line)
					continue
				}

				attribute := &Attribute{}
				attribute.AttributeName = values[0]
				if len(values) > 1 {
					attribute.Value = new(string)

					var attrValue string
					for i := 1; i < len(values); i++ {
						attrValue += values[i]
					}
					*attribute.Value = attrValue
				}
				if curMediaLevel != nil {
					curMediaLevel.Attribute = append(curMediaLevel.Attribute, attribute)
				} else {
					sdp.Attributes = append(sdp.Attributes, attribute)
				}

			}
		}
	}

	if sdp.Origin == nil {
		return nil, errors.New("no origion o=")
	}
	if sdp.SessionName == nil {
		return nil, errors.New("no session name s=")
	}
	if sdp.Timing == nil {
		return nil, errors.New("no time the session is active t=")
	}

	return
}
