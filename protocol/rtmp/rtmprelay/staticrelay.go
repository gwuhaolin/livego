package rtmprelay

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"livego/av"
	"livego/configure"
	"livego/protocol/rtmp/core"
)

type StaticPush struct {
	RtmpUrl       string
	packet_chan   chan *av.Packet
	sndctrl_chan  chan string
	connectClient *core.ConnClient
	startflag     bool
}

var G_StaticPushMap = make(map[string](*StaticPush))
var g_MapLock = new(sync.RWMutex)

var (
	STATIC_RELAY_STOP_CTRL = "STATIC_RTMPRELAY_STOP"
)

func GetStaticPushList(appname string) ([]string, error) {
	pushurlList, ok := configure.GetStaticPushUrlList(appname)

	if !ok {
		return nil, errors.New("no static push url")
	}

	return pushurlList, nil
}

func GetAndCreateStaticPushObject(rtmpurl string) *StaticPush {
	g_MapLock.RLock()
	staticpush, ok := G_StaticPushMap[rtmpurl]
	log.Printf("GetAndCreateStaticPushObject: %s, return %v", rtmpurl, ok)
	if !ok {
		g_MapLock.RUnlock()
		newStaticpush := NewStaticPush(rtmpurl)

		g_MapLock.Lock()
		G_StaticPushMap[rtmpurl] = newStaticpush
		g_MapLock.Unlock()

		return newStaticpush
	}
	g_MapLock.RUnlock()

	return staticpush
}

func GetStaticPushObject(rtmpurl string) (*StaticPush, error) {
	g_MapLock.RLock()
	if staticpush, ok := G_StaticPushMap[rtmpurl]; ok {
		g_MapLock.RUnlock()
		return staticpush, nil
	}
	g_MapLock.RUnlock()

	return nil, errors.New(fmt.Sprintf("G_StaticPushMap[%s] not exist...."))
}

func ReleaseStaticPushObject(rtmpurl string) {
	g_MapLock.RLock()
	if _, ok := G_StaticPushMap[rtmpurl]; ok {
		g_MapLock.RUnlock()

		log.Printf("ReleaseStaticPushObject %s ok", rtmpurl)
		g_MapLock.Lock()
		delete(G_StaticPushMap, rtmpurl)
		g_MapLock.Unlock()
	} else {
		g_MapLock.RUnlock()
		log.Printf("ReleaseStaticPushObject: not find %s", rtmpurl)
	}
}

func NewStaticPush(rtmpurl string) *StaticPush {
	return &StaticPush{
		RtmpUrl:       rtmpurl,
		packet_chan:   make(chan *av.Packet, 500),
		sndctrl_chan:  make(chan string),
		connectClient: nil,
		startflag:     false,
	}
}

func (self *StaticPush) Start() error {
	if self.startflag {
		return errors.New(fmt.Sprintf("StaticPush already start %s", self.RtmpUrl))
	}

	self.connectClient = core.NewConnClient()

	log.Printf("static publish server addr:%v starting....", self.RtmpUrl)
	err := self.connectClient.Start(self.RtmpUrl, "publish")
	if err != nil {
		log.Printf("connectClient.Start url=%v error", self.RtmpUrl)
		return err
	}
	log.Printf("static publish server addr:%v started, streamid=%d", self.RtmpUrl, self.connectClient.GetStreamId())
	go self.HandleAvPacket()

	self.startflag = true
	return nil
}

func (self *StaticPush) Stop() {
	if !self.startflag {
		return
	}

	log.Printf("StaticPush Stop: %s", self.RtmpUrl)
	self.sndctrl_chan <- STATIC_RELAY_STOP_CTRL
	self.startflag = false
}

func (self *StaticPush) WriteAvPacket(packet *av.Packet) {
	if !self.startflag {
		return
	}

	self.packet_chan <- packet
}

func (self *StaticPush) sendPacket(p *av.Packet) {
	if !self.startflag {
		return
	}
	var cs core.ChunkStream

	cs.Data = p.Data
	cs.Length = uint32(len(p.Data))
	cs.StreamID = self.connectClient.GetStreamId()
	cs.Timestamp = p.TimeStamp
	//cs.Timestamp += v.BaseTimeStamp()

	//log.Printf("Static sendPacket: rtmpurl=%s, length=%d, streamid=%d",
	//	self.RtmpUrl, len(p.Data), cs.StreamID)
	if p.IsVideo {
		cs.TypeID = av.TAG_VIDEO
	} else {
		if p.IsMetadata {
			cs.TypeID = av.TAG_SCRIPTDATAAMF0
		} else {
			cs.TypeID = av.TAG_AUDIO
		}
	}

	self.connectClient.Write(cs)
}

func (self *StaticPush) HandleAvPacket() {
	if !self.IsStart() {
		log.Printf("static push %s not started", self.RtmpUrl)
		return
	}

	for {
		select {
		case packet := <-self.packet_chan:
			self.sendPacket(packet)
		case ctrlcmd := <-self.sndctrl_chan:
			if ctrlcmd == STATIC_RELAY_STOP_CTRL {
				self.connectClient.Close(nil)
				log.Printf("Static HandleAvPacket close: publishurl=%s", self.RtmpUrl)
				break
			}
		}
	}
}

func (self *StaticPush) IsStart() bool {
	return self.startflag
}
