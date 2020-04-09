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
	redisAddr = flag.String("redis_addr", "", "redis addr to save room keys ex. localhost:6379")
	redisPwd  = flag.String("redis_pwd", "", "redis password")
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
	RedisAddr string `json:"redis_addr"`
	RedisPwd  string `json:"redis_pwd"`
	JWTCfg    `json:"jwt"`
	Server    []Application `json:"server"`
}

var RtmpServercfg ServerCfg

func LoadConfig(configfilename string) error {
	log.Printf("starting load configure file %s", configfilename)
	data, err := ioutil.ReadFile(configfilename)
	if err != nil {
		log.Printf("ReadFile %s error:%v", configfilename, err)
		return err
	}

	// log.Printf("loadconfig: \r\n%s", string(data))

	err = json.Unmarshal(data, &RtmpServercfg)
	if err != nil {
		log.Printf("json.Unmarshal error:%v", err)
		return err
	}
	log.Printf("get config json data:%v", RtmpServercfg)

	Init()

	return nil
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
