package connection

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gwuhaolin/livego/auth/utils"
)

type Session struct {
	_Key      string
	TimeStamp string `json:"time,omitempty"`
	Code      string `json:"code,omitempty"`
	IpAddress string `json:"ip,omitempty"`
	Active    bool   `json:"active,omitempty"`
}

type Connections = []Session

func New(code, IpAddress string) *Session {
	return &Session{
		_Key:      fmt.Sprintf("session-[%s-%s]", utils.RandomNumberString(4), utils.RandomNumberString(5)),
		TimeStamp: time.Now().UTC().Format(time.RFC3339Nano),
		Code:      code,
		IpAddress: IpAddress,
		Active:    true,
	}
}

func (session *Session) Key() string {
	return session._Key
}

func (client *Session) ToJson() ([]byte, error) {
	return json.Marshal(client)
}

func FromJson(data []byte) (conn *Session, err error) {
	conn = &Session{}
	err = json.Unmarshal(data, conn)
	return
}
