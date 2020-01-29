package configure

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"time"
)

const roomKeySaveFile = "room_keys.json"

var RoomKeys = LoadRoomKey(roomKeySaveFile)

var roomUpdated = false

func init() {
	rand.Seed(time.Now().UnixNano())
	go func() {
		for {
			time.Sleep(15 * time.Second)
			if roomUpdated {
				RoomKeys.Save(roomKeySaveFile)
				roomUpdated = false
			}
		}
	}()
}


type RoomKeysType struct {
	mapChanKey sync.Map
	mapKeyChan sync.Map
}

func LoadRoomKey(f string) *RoomKeysType {
	result := &RoomKeysType {
		mapChanKey: sync.Map{},
		mapKeyChan: sync.Map{},
	}
	raw := map[string]string{}
	content, err := ioutil.ReadFile(f)
	if err != nil {
		log.Printf("Failed to read file %s for room keys", f)
		return result
	}
	if json.Unmarshal(content, &raw) != nil {
		log.Printf("Failed to unmarshal file %s for room keys", f)
		return result
	}
	for room, key := range raw {
		result.mapChanKey.Store(room, key)
		result.mapKeyChan.Store(key, room)
	}
	return result
}

func (r *RoomKeysType) Save(f string) {
	raw := map[string]string{}
	r.mapChanKey.Range(func(channel, key interface{}) bool {
		raw[channel.(string)] = key.(string)
		return true
	})
	content, err := json.Marshal(raw)
	if err != nil {
		log.Println("Failed to marshal room keys")
		return
	}
	if ioutil.WriteFile(f, content, 0644) != nil {
		log.Println("Failed to save room keys")
		return
	}
}

// set/reset a random key for channel
func (r *RoomKeysType) SetKey(channel string) string {
	var key string
	for {
		key = randStringRunes(48)
		if _, found := r.mapKeyChan.Load(key); !found {
			r.mapChanKey.Store(channel, key)
			r.mapKeyChan.Store(key, channel)
			break
		}
	}
	roomUpdated = true
	return key
}

func (r *RoomKeysType) GetKey(channel string) string {
	var key interface{}
	var found bool
	if key, found = r.mapChanKey.Load(channel); found {
		return key.(string)
	} else {
		newkey := r.SetKey(channel)
		log.Printf("[KEY] new channel [%s]: %s", channel, newkey)
		return newkey
	}
}

func (r *RoomKeysType) GetChannel(key string) string {
	channel, found := r.mapKeyChan.Load(key)
	if found {
		return channel.(string)
	} else {
		return ""
	}
}

func (r *RoomKeysType) DeleteChannel(channel string) bool {
	key, ok := r.mapChanKey.Load(channel)
	if ok {
		r.mapChanKey.Delete(channel)
		r.mapKeyChan.Delete(key)
		return true
	}
	return false
}

func (r *RoomKeysType) DeleteKey(key string) bool {
	channel, ok := r.mapKeyChan.Load(key)
	if ok {
		r.mapChanKey.Delete(channel)
		r.mapKeyChan.Delete(key)
		return true
	}
	return false
}


// helpers
var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
