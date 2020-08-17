package rtsp

import (
	"log"
)

//Client RTSP client
type Client struct {
	url string
}

//Dial connect to RTSP server
func (client *Client) Dial() (err error) {
	addr, path, err := parseURL(client.url)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(addr)
	log.Println(path)
	return
}
