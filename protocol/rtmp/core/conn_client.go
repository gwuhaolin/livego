package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	neturl "net/url"
	"strings"

	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/gwuhaolin/livego/av"
	"log"
)

var (
	respResult     = "_result"
	respError      = "_error"
	onStatus       = "onStatus"
	publishStart   = "NetStream.Publish.Start"
	playStart      = "NetStream.Play.Start"
	connectSuccess = "NetConnection.Connect.Success"
)

var (
	ErrFail = errors.New("respone err")
)

type ConnClient struct {
	done       bool
	transID    int
	url        string
	tcurl      string
	app        string
	title      string
	query      string
	curcmdName string
	streamid   uint32
	conn       *Conn
	encoder    *amf.Encoder
	decoder    *amf.Decoder
	bytesw     *bytes.Buffer
}

func NewConnClient() *ConnClient {
	return &ConnClient{
		transID: 1,
		bytesw:  bytes.NewBuffer(nil),
		encoder: &amf.Encoder{},
		decoder: &amf.Decoder{},
	}
}

func (self *ConnClient) readRespMsg() error {
	var err error
	var rc ChunkStream
	for {
		if err = self.conn.Read(&rc); err != nil {
			return err
		}
		switch rc.TypeID {
		case 20, 17:
			r := bytes.NewReader(rc.Data)
			vs, err := self.decoder.DecodeBatch(r, amf.AMF0)
			if err != nil && err != io.EOF {
				return err
			}
			for k, v := range vs {
				switch v.(type) {
				case string:
					switch self.curcmdName {
					case cmdConnect, cmdCreateStream:
						if v.(string) != respResult {
							return ErrFail
						}
					case cmdPublish:
						if v.(string) != onStatus {
							return ErrFail
						}
					}
				case float64:
					switch self.curcmdName {
					case cmdConnect, cmdCreateStream:
						id := int(v.(float64))
						if k == 1 {
							if id != self.transID {
								return ErrFail
							}
						} else if k == 3 {
							self.streamid = uint32(id)
						}
					case cmdPublish:
						if int(v.(float64)) != 0 {
							return ErrFail
						}
					}
				case amf.Object:
					objmap := v.(amf.Object)
					switch self.curcmdName {
					case cmdConnect:
						code, ok := objmap["code"]
						if ok && code.(string) != connectSuccess {
							return ErrFail
						}
					case cmdPublish:
						code, ok := objmap["code"]
						if ok && code.(string) != publishStart {
							return ErrFail
						}
					}
				}
			}
			return nil
		}
	}
}

func (self *ConnClient) writeMsg(args ...interface{}) error {
	self.bytesw.Reset()
	for _, v := range args {
		if _, err := self.encoder.Encode(self.bytesw, v, amf.AMF0); err != nil {
			return err
		}
	}
	msg := self.bytesw.Bytes()
	c := ChunkStream{
		Format:    0,
		CSID:      3,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  self.streamid,
		Length:    uint32(len(msg)),
		Data:      msg,
	}
	self.conn.Write(&c)
	return self.conn.Flush()
}

func (self *ConnClient) writeConnectMsg() error {
	event := make(amf.Object)
	event["app"] = self.app
	event["type"] = "nonprivate"
	event["flashVer"] = "FMS.3.1"
	event["tcUrl"] = self.tcurl
	self.curcmdName = cmdConnect

	if err := self.writeMsg(cmdConnect, self.transID, event); err != nil {
		return err
	}
	return self.readRespMsg()
}

func (self *ConnClient) writeCreateStreamMsg() error {
	self.transID++
	self.curcmdName = cmdCreateStream
	if err := self.writeMsg(cmdCreateStream, self.transID, nil); err != nil {
		return err
	}
	return self.readRespMsg()
}

func (self *ConnClient) writePublishMsg() error {
	self.transID++
	self.curcmdName = cmdPublish
	if err := self.writeMsg(cmdPublish, self.transID, nil, self.title, publishLive); err != nil {
		return err
	}
	return self.readRespMsg()
}

func (self *ConnClient) writePlayMsg() error {
	self.transID++
	self.curcmdName = cmdPlay
	if err := self.writeMsg(cmdPlay, 0, nil, self.title); err != nil {
		return err
	}
	return self.readRespMsg()
}

func (self *ConnClient) Start(url string, method string) error {
	u, err := neturl.Parse(url)
	if err != nil {
		return err
	}
	self.url = url
	path := strings.TrimLeft(u.Path, "/")
	ps := strings.SplitN(path, "/", 2)
	if len(ps) != 2 {
		return fmt.Errorf("u path err: %s", path)
	}
	self.app = ps[0]
	self.title = ps[1]
	self.query = u.RawQuery
	self.tcurl = "rtmp://" + u.Host + "/" + self.app
	port := ":1935"
	host := u.Host
	localIP := ":0"
	var remoteIP string
	if strings.Index(host, ":") != -1 {
		host, port, err = net.SplitHostPort(host)
		if err != nil {
			return err
		}
		port = ":" + port
	}
	ips, err := net.LookupIP(host)
	log.Printf("ips: %v, host: %v", ips, host)
	if err != nil {
		log.Println(err)
		return err
	}
	remoteIP = ips[rand.Intn(len(ips))].String()
	if strings.Index(remoteIP, ":") == -1 {
		remoteIP += port
	}

	local, err := net.ResolveTCPAddr("tcp", localIP)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("remoteIP: ", remoteIP)
	remote, err := net.ResolveTCPAddr("tcp", remoteIP)
	if err != nil {
		log.Println(err)
		return err
	}
	conn, err := net.DialTCP("tcp", local, remote)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("connection:", "local:", conn.LocalAddr(), "remote:", conn.RemoteAddr())

	self.conn = NewConn(conn, 4*1024)
	if err := self.conn.HandshakeClient(); err != nil {
		return err
	}
	if err := self.writeConnectMsg(); err != nil {
		return err
	}
	if err := self.writeCreateStreamMsg(); err != nil {
		return err
	}
	if method == av.PUBLISH {
		if err := self.writePublishMsg(); err != nil {
			return err
		}
	} else if method == av.PLAY {
		if err := self.writePlayMsg(); err != nil {
			return err
		}
	}

	return nil
}

func (self *ConnClient) Write(c ChunkStream) error {
	if c.TypeID == av.TAG_SCRIPTDATAAMF0 ||
		c.TypeID == av.TAG_SCRIPTDATAAMF3 {
		var err error
		if c.Data, err = amf.MetaDataReform(c.Data, amf.ADD); err != nil {
			return err
		}
		c.Length = uint32(len(c.Data))
	}
	return self.conn.Write(&c)
}

func (self *ConnClient) Read(c *ChunkStream) (err error) {
	return self.conn.Read(c)
}

func (self *ConnClient) GetInfo() (app string, name string, url string) {
	app = self.app
	name = self.title
	url = self.url
	return
}

func (self *ConnClient) Close(err error) {
	self.conn.Close()
}
