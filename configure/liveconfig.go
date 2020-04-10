package configure

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

/*
{
  "server": [
    {
      "appname": "live",
      "liveon": "on",
	  "hlson": "on",
	  "static_push": []
    }
  ]
}
*/
var (
	roomKeySaveFile = flag.String("KeyFile", "room_keys.json", "path to save room keys")
	RedisAddr       = flag.String("redis_addr", "", "redis addr to save room keys ex. localhost:6379")
	RedisPwd        = flag.String("redis_pwd", "", "redis password")
)

type Application struct {
	Appname    string   `json:"appname"`
	Liveon     string   `json:"liveon"`
	Hlson      string   `json:"hlson"`
	StaticPush []string `json:"static_push"`
}

type JWTCfg struct {
	Secret    string `json:"secret"`
	Algorithm string `json:"algorithm"`
}

type ServerCfg struct {
	KeyFile   string `json:"key_file"`
	RedisAddr string `json:"redis_addr"`
	RedisPwd  string `json:"redis_pwd"`
	JWTCfg    `json:"jwt"`
	Server    []Application `json:"server"`
}

// default config
var RtmpServercfg = ServerCfg{
	Server: []Application{{
		Appname:    "livego",
		Liveon:     "on",
		Hlson:      "on",
		StaticPush: nil,
	}},
}

func LoadConfig(configfilename string) {
	log.Printf("starting load configure file %s", configfilename)
	data, err := ioutil.ReadFile(configfilename)
	if err != nil {
		log.Printf("ReadFile %s error:%v", configfilename, err)
	}

	err = json.Unmarshal(data, &RtmpServercfg)
	if err != nil {
		log.Printf("json.Unmarshal error:%v", err)
	}
	log.Printf("get config json data:%v", RtmpServercfg)

	Init()
}

func GetKeyFile() *string {
	if len(RtmpServercfg.KeyFile) > 0 {
		*roomKeySaveFile = RtmpServercfg.KeyFile
	}

	return roomKeySaveFile
}

func GetRedisAddr() *string {
	if len(RtmpServercfg.RedisAddr) > 0 {
		*RedisAddr = RtmpServercfg.RedisAddr
	}

	if len(*RedisAddr) == 0 {
		return nil
	}

	return RedisAddr
}

func GetRedisPwd() *string {
	if len(RtmpServercfg.RedisPwd) > 0 {
		*RedisPwd = RtmpServercfg.RedisPwd
	}

	return RedisPwd
}

func CheckAppName(appname string) bool {
	for _, app := range RtmpServercfg.Server {
		if (app.Appname == appname) && (app.Liveon == "on") {
			return true
		}
	}
	return false
}

func GetStaticPushUrlList(appname string) ([]string, bool) {
	for _, app := range RtmpServercfg.Server {
		if (app.Appname == appname) && (app.Liveon == "on") {
			if len(app.StaticPush) > 0 {
				return app.StaticPush, true
			} else {
				return nil, false
			}
		}

	}
	return nil, false
}
