package client

import "encoding/json"

type Client struct {
	Email          string `json:"email,omitempty"`
	Code           string `json:"code,omitempty"`
	Key            string `json:"key,omitempty"`
	Active         bool   `json:"active,omitempty"`
	LastSessionKey string `json:"lsk,omitempty"`
}

type Clients = []Client

func New() *Client {
	return &Client{}
}

func (client *Client) ToJson() ([]byte, error) {
	return json.Marshal(client)
}

func FromJson(data []byte) (client *Client, err error) {
	client = New()
	err = json.Unmarshal(data, client)
	return
}
