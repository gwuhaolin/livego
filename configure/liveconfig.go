package configure

import (
	"encoding/json"
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
	JWTCfg `json:"jwt"`
	Server []Application `json:"server"`
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
	return nil
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
