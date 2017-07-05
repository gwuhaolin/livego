package configure

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

/*
{
	[
	{
	"application":"live",
	"live":"on",
	"hls":"on",
	"static_push":["rtmp://xx/live"]
	}
	]
}
*/
type Application struct {
	Appname     string
	Liveon      string
	Hlson       string
	Static_push []string
}

type ServerCfg struct {
	Server []Application
}

var RtmpServercfg ServerCfg

func LoadConfig(configfilename string) error {
	log.Printf("starting load configure file(%s)......", configfilename)
	data, err := ioutil.ReadFile(configfilename)
	if err != nil {
		log.Printf("ReadFile %s error:%v", configfilename, err)
		return err
	}

	log.Printf("loadconfig: \r\n%s", string(data))

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
			if len(app.Static_push) > 0 {
				return app.Static_push, true
			} else {
				return nil, false
			}
		}

	}
	return nil, false
}
