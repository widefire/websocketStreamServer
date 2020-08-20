package rtsp

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

//TransportSpec ...
type TransportSpec struct {
	TransportProtocol string //    "RTP"
	Profile           string //  "AVP"
	LowerTransport    string // "TCP" | "UDP"
}

//ParseTransportSpec ...
func (trans *TransportSpec) ParseTransportSpec(str string) (err error) {
	sub := strings.SplitN(str, "/", 3)
	if len(sub) < 2 {
		err = fmt.Errorf("bad transport spec %s", str)
		log.Println(err)
		return
	}
	trans.TransportProtocol = sub[0]
	trans.Profile = sub[1]
	if len(sub) == 3 {
		trans.LowerTransport = sub[2]
	} else {
		trans.LowerTransport = "UDP"
	}
	return
}

//IntRange ...
type IntRange struct {
	From int
	To   *int
}

func createIntRange(str string) (ir *IntRange, err error) {
	sub := strings.SplitN(str, "-", 2)
	ir = &IntRange{}
	ir.From, err = strconv.Atoi(sub[0])
	if err != nil {
		log.Println(err)
		return
	}
	if len(sub) > 1 {
		ir.To = new(int)
		*ir.To, err = strconv.Atoi(sub[1])
		if err != nil {
			log.Println(err)
			return
		}
	}
	return
}

//TransportItem ...
type TransportItem struct {
	TransportSpec
	//params
	Unicast     bool      //"unicast" | "multicast"
	Destination string    //"destination"
	Interleaved *IntRange //"interleaved"
	Append      bool      //"append"
	TTL         int       //"ttl"
	Layers      int       //"layers"
	Port        *IntRange //"port"
	ClientPort  *IntRange //"client_port"
	ServerPort  *IntRange //"server_port"
	SSRC        string    //"ssrc"
	Mode        string    //"mode"
}

//Transport ...
type Transport struct {
	Items []*TransportItem
}

//ParseTransport ...
func ParseTransport(header http.Header) (trans *Transport, err error) {
	trans = &Transport{}
	trans.Items = make([]*TransportItem, 0)
	value := header.Get("Transport")
	if len(value) == 0 {
		return
	}
	strItems := strings.Split(value, ",")
	for _, strItem := range strItems {
		item := &TransportItem{}
		strProps := strings.Split(strItem, ";")
		if len(strProps) == 0 {
			log.Println(strItem)
			continue
		}
		err = item.ParseTransportSpec(strProps[0])
		if err != nil {
			log.Println(err)
			return
		}
		for i := 1; i < len(strProps); i++ {
			prop := strProps[i]

			if prop == "unicast" {
				item.Unicast = true
			} else if prop == "multicast" {
				item.Unicast = false
			} else if prop == "append" {
				item.Append = true
			} else {
				kvs := strings.SplitN(prop, "=", 2)
				if len(kvs) == 2 {
					if len(kvs[1]) > 0 {
						switch kvs[0] {
						case "destination":
							item.Destination = kvs[1]
						case "interleaved":
							item.Interleaved, err = createIntRange(kvs[1])
							if err != nil {
								log.Println(err)
								return
							}
						case "ttl":
							item.TTL, err = strconv.Atoi(kvs[1])
							if err != nil {
								log.Println(err)
								return
							}
						case "layers":
							item.Layers, err = strconv.Atoi(kvs[1])
							if err != nil {
								log.Println(err)
								return
							}
						case "port":
							item.Port, err = createIntRange(kvs[1])
							if err != nil {
								log.Println(err)
								return
							}
						case "client_port":
							item.ClientPort, err = createIntRange(kvs[1])
							if err != nil {
								log.Println(err)
								return
							}
						case "server_port":
							item.ServerPort, err = createIntRange(kvs[1])
							if err != nil {
								log.Println(err)
								return
							}
						case "ssrc":
							item.SSRC = kvs[1]
						case "mode":
							item.Mode = kvs[1]
						}
					} else {
						log.Printf("empty prop value %s\r\n", prop)
					}
				} else {
					log.Printf("invalid prop %s", prop)
				}
			}
		}
		trans.Items = append(trans.Items, item)
	}
	return
}
