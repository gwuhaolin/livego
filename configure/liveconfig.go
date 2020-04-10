package configure

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

/*
{
  "server": [
    {
      "appname": "live",
      "live": true,
	  "hls": true,
	  "static_push": []
    }
  ]
}
*/
var (
	redisAddr = flag.String("redis_addr", "", "redis addr to save room keys ex. localhost:6379")
	redisPwd  = flag.String("redis_pwd", "", "redis password")
)

type Application struct {
	Appname    string   `json:"appname"`
	Live       bool     `json:"liveon"`
	Hls        bool     `json:"hls"`
	StaticPush []string `json:"static_push"`
}
type JWTCfg struct {
	Secret    string `json:"secret"`
	Algorithm string `json:"algorithm"`
}
type ServerCfg struct {
	RedisAddr string `json:"redis_addr"`
	RedisPwd  string `json:"redis_pwd"`
	JWTCfg    `json:"jwt"`
	Server    []Application `json:"server"`
}

// default config
var RtmpServercfg = ServerCfg{
	Server: []Application{{
		Appname:    "livego",
		Live:       true,
		Hls:        true,
		StaticPush: nil,
	}},
}

func LoadConfig(configfilename string) {
	defer Init()

	log.Infof("starting load configure file %s", configfilename)
	data, err := ioutil.ReadFile(configfilename)
	if err != nil {
		log.Warningf("ReadFile %s error:%v", configfilename, err)
		log.Info("Using default config")
		return
	}

	err = json.Unmarshal(data, &RtmpServercfg)
	if err != nil {
		log.Errorf("json.Unmarshal error:%v", err)
		log.Info("Using default config")
	}
	log.Debugf("get config json data:%v", RtmpServercfg)
}

func GetRedisAddr() *string {
	if len(RtmpServercfg.RedisAddr) > 0 {
		*redisAddr = RtmpServercfg.RedisAddr
	}

	if len(*redisAddr) == 0 {
		return nil
	}

	return redisAddr
}

func GetRedisPwd() *string {
	if len(RtmpServercfg.RedisPwd) > 0 {
		*redisPwd = RtmpServercfg.RedisPwd
	}

	return redisPwd
}

func CheckAppName(appname string) bool {
	for _, app := range RtmpServercfg.Server {
		if app.Appname == appname {
			return app.Live
		}
	}
	return false
}

func GetStaticPushUrlList(appname string) ([]string, bool) {
	for _, app := range RtmpServercfg.Server {
		if (app.Appname == appname) && app.Live {
			if len(app.StaticPush) > 0 {
				return app.StaticPush, true
			} else {
				return nil, false
			}
		}

	}
	return nil, false
}
