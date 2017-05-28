package main

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/gwuhaolin/livego/protocol/rtmp"
	"github.com/gwuhaolin/livego/protocol/hls"
	"github.com/gwuhaolin/livego/protocol/httpflv"
	"github.com/gwuhaolin/livego/protocol/httpopera"
	"path/filepath"
	"strings"
	"io/ioutil"
	"strconv"
	"log"
)

var (
	rtmpAddr  = flag.String("rtmpAddr", ":1935", "The rtmp server address to bind.")
	flvAddr   = flag.String("flvAddr", ":8081", "the http-flv server address to bind.")
	hlsAddr   = flag.String("hlsAddr", ":8080", "the hls server address to bind.")
	operaAddr = flag.String("operaAddr", "", "the http operation or config address to bind: 8082.")
	CurDir    string // save pid
)

func getParentDirectory(dirctory string) string {
	return substr(dirctory, 0, strings.LastIndex(dirctory, "/"))
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

func SavePid() error {
	pidFilename := CurDir + "/pid/" + filepath.Base(os.Args[0]) + ".pid"
	pid := os.Getpid()
	return ioutil.WriteFile(pidFilename, []byte(strconv.Itoa(pid)), 0755)
}

func init() {
	CurDir = getParentDirectory(getCurrentDirectory())
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()
}

func catchSignal() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGSTOP, syscall.SIGTERM)
	<-sig
	log.Println("recieved signal!")
}

func startHls() *hls.Server {
	hlsListen, err := net.Listen("tcp", *hlsAddr)
	if err != nil {
		log.Fatal(err)
	}

	hlsServer := hls.NewServer()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("hls server panic: ", r)
			}
		}()
		hlsServer.Serve(hlsListen)
	}()
	return hlsServer
}

func startRtmp(stream *rtmp.RtmpStream, hlsServer *hls.Server) {
	rtmplisten, err := net.Listen("tcp", *rtmpAddr)
	if err != nil {
		log.Fatal(err)
	}

	rtmpServer := rtmp.NewRtmpServer(stream, hlsServer)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("hls server panic: ", r)
			}
		}()
		rtmpServer.Serve(rtmplisten)
	}()
}

func startHTTPFlv(stream *rtmp.RtmpStream) {
	flvListen, err := net.Listen("tcp", *flvAddr)
	if err != nil {
		log.Fatal(err)
	}

	hdlServer := httpflv.NewServer(stream)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("hls server panic: ", r)
			}
		}()
		hdlServer.Serve(flvListen)
	}()
}

func startHTTPOpera(stream *rtmp.RtmpStream) {
	if *operaAddr != "" {
		opListen, err := net.Listen("tcp", *operaAddr)
		if err != nil {
			log.Fatal(err)
		}
		opServer := httpopera.NewServer(stream)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("hls server panic: ", r)
				}
			}()
			opServer.Serve(opListen)
		}()
	}
}

func startLog() {
	log.Println("RTMP Listen On", *rtmpAddr)
	log.Println("HLS Listen On", *hlsAddr)
	log.Println("HTTP-FLV Listen On", *flvAddr)
	if *operaAddr != "" {
		log.Println("HTTP-Operation Listen On", *operaAddr)
	}
	SavePid()
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("main panic: ", r)
			time.Sleep(1 * time.Second)
		}
	}()

	stream := rtmp.NewRtmpStream()
	// hls
	h := startHls()
	// rtmp
	startRtmp(stream, h)
	// http-flv
	startHTTPFlv(stream)
	// http-opera
	startHTTPOpera(stream)
	// my log
	startLog()
	// block
	catchSignal()
}
