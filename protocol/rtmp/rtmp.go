package rtmp

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/ClipMyHorseTV/livego/utils/uid"

	"github.com/ClipMyHorseTV/livego/av"
	"github.com/ClipMyHorseTV/livego/configure"
	"github.com/ClipMyHorseTV/livego/container/flv"
	"github.com/ClipMyHorseTV/livego/protocol/rtmp/core"

	log "github.com/sirupsen/logrus"
)

const (
	maxQueueNum           = 1024
	SAVE_STATICS_INTERVAL = 5000
)

var (
	readTimeout  = configure.Config.GetInt("read_timeout")
	writeTimeout = configure.Config.GetInt("write_timeout")
)

type Client struct {
	handler av.Handler
	getter  av.GetWriter
}

func NewRtmpClient(h av.Handler, getter av.GetWriter) *Client {
	return &Client{
		handler: h,
		getter:  getter,
	}
}

func (c *Client) Dial(url string, method string) error {
	connClient := core.NewConnClient()
	if err := connClient.Start(url, method); err != nil {
		return err
	}
	if method == av.PUBLISH {
		writer := NewVirWriter(connClient)
		log.Debugf("client Dial call NewVirWriter url=%s, method=%s", url, method)
		c.handler.HandleWriter(writer)
	} else if method == av.PLAY {
		reader := NewVirReader(connClient)
		log.Debugf("client Dial call NewVirReader url=%s, method=%s", url, method)
		c.handler.HandleReader(reader)
		if c.getter != nil {
			writer := c.getter.GetWriter(reader.Info())
			c.handler.HandleWriter(writer)
		}
	}
	return nil
}

func (c *Client) GetHandle() av.Handler {
	return c.handler
}

type Server struct {
	handler av.Handler
	getter  av.GetWriter
}

func NewRtmpServer(h av.Handler, getter av.GetWriter) *Server {
	return &Server{
		handler: h,
		getter:  getter,
	}
}

func (s *Server) Serve(listener net.Listener) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("rtmp serve panic: ", r)
		}
	}()

	for {
		var netconn net.Conn
		netconn, err = listener.Accept()
		if err != nil {
			return
		}
		conn := core.NewConn(netconn, 4*1024)
		log.Debug("new client, connect remote: ", conn.RemoteAddr().String(),
			"local:", conn.LocalAddr().String())
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn *core.Conn) error {
	if err := conn.HandshakeServer(); err != nil {
		conn.Close()
		log.Error("handleConn HandshakeServer err: ", err)
		return err
	}
	connServer := core.NewConnServer(conn)

	if err := connServer.ReadMsg(); err != nil {
		conn.Close()
		log.Error("handleConn read msg err: ", err)
		return err
	}

	appname, name, _ := connServer.GetInfo()

	if ret := configure.CheckAppName(appname); !ret {
		err := fmt.Errorf("application name=%s is not configured", appname)
		conn.Close()
		log.Error("CheckAppName err: ", err)
		return err
	}

	log.Debugf("handleConn: IsPublisher=%v", connServer.IsPublisher())
	if connServer.IsPublisher() {
		if configure.Config.GetBool("rtmp_noauth") {
			key, err := configure.RoomKeys.GetKey(name)
			if err != nil {
				err := fmt.Errorf("Cannot create key err=%s", err.Error())
				conn.Close()
				log.Error("GetKey err: ", err)
				return err
			}
			name = key
		}
		channel, err := configure.RoomKeys.GetChannel(name)
		if err != nil {
			err := fmt.Errorf("invalid key err=%s", err.Error())
			conn.Close()
			log.Error("CheckKey err: ", err)
			return err
		}
		connServer.PublishInfo.Name = channel
		if pushlist, ret := configure.GetStaticPushUrlList(appname); ret && (pushlist != nil) {
			log.Debugf("GetStaticPushUrlList: %v", pushlist)
		}
		reader := NewVirReader(connServer)
		s.handler.HandleReader(reader)
		log.Debugf("new publisher: %+v", reader.Info())

		if s.getter != nil {
			writeType := reflect.TypeOf(s.getter)
			log.Debugf("handleConn:writeType=%v", writeType)
			writer := s.getter.GetWriter(reader.Info())
			s.handler.HandleWriter(writer)
		}
		if configure.Config.GetBool("flv_archive") {
			flvWriter := new(flv.FlvDvr)
			s.handler.HandleWriter(flvWriter.GetWriter(reader.Info()))
		}
	} else {
		writer := NewVirWriter(connServer)
		log.Debugf("new player: %+v", writer.Info())
		s.handler.HandleWriter(writer)
	}

	return nil
}

type GetInFo interface {
	GetInfo() (string, string, string)
}

type StreamReadWriteCloser interface {
	GetInFo
	Close(error)
	Write(core.ChunkStream) error
	Read(c *core.ChunkStream) error
}

type StaticsBW struct {
	StreamId               uint32
	VideoDatainBytes       uint64
	LastVideoDatainBytes   uint64
	VideoSpeedInBytesperMS uint64

	AudioDatainBytes       uint64
	LastAudioDatainBytes   uint64
	AudioSpeedInBytesperMS uint64

	LastTimestamp int64
}

type VirWriter struct {
	Uid string
	av.RWBaser
	conn StreamReadWriteCloser
	// packetQueue is never closed, so concurrent Write and Close cannot race
	// on it. Shutdown is signalled through done instead.
	packetQueue chan *av.Packet
	done        chan struct{}
	closeOnce   sync.Once
	WriteBWInfo StaticsBW
}

func NewVirWriter(conn StreamReadWriteCloser) *VirWriter {
	ret := &VirWriter{
		Uid:         uid.NewId(),
		conn:        conn,
		RWBaser:     av.NewRWBaser(time.Second * time.Duration(writeTimeout)),
		packetQueue: make(chan *av.Packet, maxQueueNum),
		done:        make(chan struct{}),
		WriteBWInfo: StaticsBW{0, 0, 0, 0, 0, 0, 0, 0},
	}

	go ret.Check()
	go func() {
		err := ret.SendPacket()
		if err != nil {
			log.Warning(err)
		}
	}()
	return ret
}

func (v *VirWriter) SaveStatics(streamid uint32, length uint64, isVideoFlag bool) {
	nowInMS := int64(time.Now().UnixNano() / 1e6)

	v.WriteBWInfo.StreamId = streamid
	if isVideoFlag {
		v.WriteBWInfo.VideoDatainBytes = v.WriteBWInfo.VideoDatainBytes + length
	} else {
		v.WriteBWInfo.AudioDatainBytes = v.WriteBWInfo.AudioDatainBytes + length
	}

	if v.WriteBWInfo.LastTimestamp == 0 {
		v.WriteBWInfo.LastTimestamp = nowInMS
	} else if (nowInMS - v.WriteBWInfo.LastTimestamp) >= SAVE_STATICS_INTERVAL {
		diffTimestamp := (nowInMS - v.WriteBWInfo.LastTimestamp) / 1000

		v.WriteBWInfo.VideoSpeedInBytesperMS = (v.WriteBWInfo.VideoDatainBytes - v.WriteBWInfo.LastVideoDatainBytes) * 8 / uint64(diffTimestamp) / 1000
		v.WriteBWInfo.AudioSpeedInBytesperMS = (v.WriteBWInfo.AudioDatainBytes - v.WriteBWInfo.LastAudioDatainBytes) * 8 / uint64(diffTimestamp) / 1000

		v.WriteBWInfo.LastVideoDatainBytes = v.WriteBWInfo.VideoDatainBytes
		v.WriteBWInfo.LastAudioDatainBytes = v.WriteBWInfo.AudioDatainBytes
		v.WriteBWInfo.LastTimestamp = nowInMS
	}
}

func (v *VirWriter) Check() {
	var c core.ChunkStream
	for {
		if err := v.conn.Read(&c); err != nil {
			v.Close(err)
			return
		}
	}
}

// markClosed signals shutdown to Write and SendPacket. Safe to call more than once.
func (v *VirWriter) markClosed() {
	v.closeOnce.Do(func() { close(v.done) })
}

func (v *VirWriter) isClosed() bool {
	select {
	case <-v.done:
		return true
	default:
		return false
	}
}

// recv takes a packet from pktQue, giving up once the writer is closed.
func (v *VirWriter) recv(pktQue chan *av.Packet) (*av.Packet, bool) {
	select {
	case p := <-pktQue:
		return p, true
	case <-v.done:
		return nil, false
	}
}

// send puts a packet into pktQue, giving up once the writer is closed.
func (v *VirWriter) send(pktQue chan *av.Packet, p *av.Packet) bool {
	select {
	case pktQue <- p:
		return true
	case <-v.done:
		return false
	}
}

func (v *VirWriter) DropPacket(pktQue chan *av.Packet, info av.Info) {
	log.Warningf("[%v] packet queue max!!!", info)
	for i := 0; i < maxQueueNum-84; i++ {
		tmpPkt, ok := v.recv(pktQue)
		if !ok {
			return
		}

		// try to don't drop audio
		if tmpPkt.IsAudio {
			if len(pktQue) > maxQueueNum-2 {
				log.Debug("drop audio pkt")
				if _, ok := v.recv(pktQue); !ok {
					return
				}
			} else if !v.send(pktQue, tmpPkt) {
				return
			}
		}

		if tmpPkt.IsVideo {
			videoPkt, ok := tmpPkt.Header.(av.VideoPacketHeader)
			// dont't drop sps config and dont't drop key frame
			if ok && (videoPkt.IsSeq() || videoPkt.IsKeyFrame()) {
				if !v.send(pktQue, tmpPkt) {
					return
				}
			}
			if len(pktQue) > maxQueueNum-10 {
				log.Debug("drop video pkt")
				if _, ok := v.recv(pktQue); !ok {
					return
				}
			}
		}

	}
	log.Debug("packet queue len: ", len(pktQue))
}

func (v *VirWriter) Write(p *av.Packet) error {
	if v.isClosed() {
		return fmt.Errorf("VirWriter closed")
	}

	if len(v.packetQueue) >= maxQueueNum-24 {
		v.DropPacket(v.packetQueue, v.Info())
	} else if !v.send(v.packetQueue, p) {
		return fmt.Errorf("VirWriter closed")
	}

	return nil
}

func (v *VirWriter) SendPacket() error {
	Flush := reflect.ValueOf(v.conn).MethodByName("Flush")
	var cs core.ChunkStream
	for {
		p, ok := v.recv(v.packetQueue)
		if ok {
			cs.Data = p.Data
			cs.Length = uint32(len(p.Data))
			cs.StreamID = p.StreamID
			cs.Timestamp = p.TimeStamp
			cs.Timestamp += v.BaseTimeStamp()

			if p.IsVideo {
				cs.TypeID = av.TAG_VIDEO
			} else {
				if p.IsMetadata {
					cs.TypeID = av.TAG_SCRIPTDATAAMF0
				} else {
					cs.TypeID = av.TAG_AUDIO
				}
			}

			v.SaveStatics(p.StreamID, uint64(cs.Length), p.IsVideo)
			v.SetPreTime()
			v.RecTimeStamp(cs.Timestamp, cs.TypeID)
			err := v.conn.Write(cs)
			if err != nil {
				v.markClosed()
				return err
			}
			Flush.Call(nil)
		} else {
			return fmt.Errorf("closed")
		}

	}
}

func (v *VirWriter) Info() (ret av.Info) {
	ret.UID = v.Uid
	_, _, URL := v.conn.GetInfo()
	ret.URL = URL
	_url, err := url.Parse(URL)
	if err != nil {
		log.Warning(err)
	}
	ret.Key = strings.TrimLeft(_url.Path, "/")
	ret.Inter = true
	return
}

func (v *VirWriter) Close(err error) {
	log.Warning("player ", v.Info(), "closed: "+err.Error())
	v.markClosed()
	v.conn.Close(err)
}

type VirReader struct {
	Uid string
	av.RWBaser
	demuxer    *flv.Demuxer
	conn       StreamReadWriteCloser
	ReadBWInfo StaticsBW
}

func NewVirReader(conn StreamReadWriteCloser) *VirReader {
	return &VirReader{
		Uid:        uid.NewId(),
		conn:       conn,
		RWBaser:    av.NewRWBaser(time.Second * time.Duration(writeTimeout)),
		demuxer:    flv.NewDemuxer(),
		ReadBWInfo: StaticsBW{0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func (v *VirReader) SaveStatics(streamid uint32, length uint64, isVideoFlag bool) {
	nowInMS := int64(time.Now().UnixNano() / 1e6)

	v.ReadBWInfo.StreamId = streamid
	if isVideoFlag {
		v.ReadBWInfo.VideoDatainBytes = v.ReadBWInfo.VideoDatainBytes + length
	} else {
		v.ReadBWInfo.AudioDatainBytes = v.ReadBWInfo.AudioDatainBytes + length
	}

	if v.ReadBWInfo.LastTimestamp == 0 {
		v.ReadBWInfo.LastTimestamp = nowInMS
	} else if (nowInMS - v.ReadBWInfo.LastTimestamp) >= SAVE_STATICS_INTERVAL {
		diffTimestamp := (nowInMS - v.ReadBWInfo.LastTimestamp) / 1000

		//log.Printf("now=%d, last=%d, diff=%d", nowInMS, v.ReadBWInfo.LastTimestamp, diffTimestamp)
		v.ReadBWInfo.VideoSpeedInBytesperMS = (v.ReadBWInfo.VideoDatainBytes - v.ReadBWInfo.LastVideoDatainBytes) * 8 / uint64(diffTimestamp) / 1000
		v.ReadBWInfo.AudioSpeedInBytesperMS = (v.ReadBWInfo.AudioDatainBytes - v.ReadBWInfo.LastAudioDatainBytes) * 8 / uint64(diffTimestamp) / 1000

		v.ReadBWInfo.LastVideoDatainBytes = v.ReadBWInfo.VideoDatainBytes
		v.ReadBWInfo.LastAudioDatainBytes = v.ReadBWInfo.AudioDatainBytes
		v.ReadBWInfo.LastTimestamp = nowInMS
	}
}

func (v *VirReader) Read(p *av.Packet) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Warning("rtmp read packet panic: ", r)
		}
	}()

	v.SetPreTime()
	var cs core.ChunkStream
	for {
		err = v.conn.Read(&cs)
		if err != nil {
			return err
		}
		if cs.TypeID == av.TAG_AUDIO ||
			cs.TypeID == av.TAG_VIDEO ||
			cs.TypeID == av.TAG_SCRIPTDATAAMF0 ||
			cs.TypeID == av.TAG_SCRIPTDATAAMF3 {
			break
		}
	}

	p.IsAudio = cs.TypeID == av.TAG_AUDIO
	p.IsVideo = cs.TypeID == av.TAG_VIDEO
	p.IsMetadata = cs.TypeID == av.TAG_SCRIPTDATAAMF0 || cs.TypeID == av.TAG_SCRIPTDATAAMF3
	p.StreamID = cs.StreamID
	p.Data = cs.Data
	p.TimeStamp = cs.Timestamp

	v.SaveStatics(p.StreamID, uint64(len(p.Data)), p.IsVideo)
	v.demuxer.DemuxH(p)
	return err
}

func (v *VirReader) Info() (ret av.Info) {
	ret.UID = v.Uid
	_, _, URL := v.conn.GetInfo()
	ret.URL = URL
	_url, err := url.Parse(URL)
	if err != nil {
		log.Warning(err)
	}
	ret.Key = strings.TrimLeft(_url.Path, "/")
	return
}

func (v *VirReader) Close(err error) {
	log.Debug("publisher ", v.Info(), "closed: "+err.Error())
	v.conn.Close(err)
}
