package configure

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
)

var RoomKeys = LoadRoomKey(*GetKeyFile())

var roomUpdated = false

var saveInFile = true
var redisCli *redis.Client

func Init() {
	saveInFile = GetRedisAddr() == nil

	rand.Seed(time.Now().UnixNano())
	if saveInFile {
		go func() {
			for {
				time.Sleep(15 * time.Second)
				if roomUpdated {
					RoomKeys.Save(*roomKeySaveFile)
					roomUpdated = false
				}
			}
		}()

		return
	}

	redisCli = redis.NewClient(&redis.Options{
		Addr:     *GetRedisAddr(),
		Password: *GetRedisPwd(),
		DB:       0,
	})

	_, err := redisCli.Ping().Result()
	if err != nil {
		panic(err)
	}

	log.Printf("Redis connected")
}

type RoomKeysType struct {
	mapChanKey sync.Map
	mapKeyChan sync.Map
}

func LoadRoomKey(f string) *RoomKeysType {
	result := &RoomKeysType{
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
func (r *RoomKeysType) SetKey(channel string) (key string, err error) {
	if !saveInFile {
		for {
			key = randStringRunes(48)
			if _, err = redisCli.Get(key).Result(); err == redis.Nil {
				err = redisCli.Set(channel, key, 0).Err()
				if err != nil {
					return
				}

				err = redisCli.Set(key, channel, 0).Err()
				return
			} else if err != nil {
				return
			}
		}
	}

	for {
		key = randStringRunes(48)
		if _, found := r.mapKeyChan.Load(key); !found {
			r.mapChanKey.Store(channel, key)
			r.mapKeyChan.Store(key, channel)
			break
		}
	}
	roomUpdated = true
	return
}

func (r *RoomKeysType) GetKey(channel string) (newKey string, err error) {
	if !saveInFile {
		if newKey, err = redisCli.Get(channel).Result(); err == redis.Nil {
			newKey, err = r.SetKey(channel)
			log.Printf("[KEY] new channel [%s]: %s", channel, newKey)
			return
		}

		return
	}

	var key interface{}
	var found bool
	if key, found = r.mapChanKey.Load(channel); found {
		return key.(string), nil
	}
	newKey, err = r.SetKey(channel)
	log.Printf("[KEY] new channel [%s]: %s", channel, newKey)
	return
}

func (r *RoomKeysType) GetChannel(key string) (channel string, err error) {
	if !saveInFile {
		return redisCli.Get(key).Result()
	}

	chann, found := r.mapKeyChan.Load(key)
	if found {
		return chann.(string), nil
	} else {
		return "", fmt.Errorf("%s does not exists", key)
	}
}

func (r *RoomKeysType) DeleteChannel(channel string) bool {
	if !saveInFile {
		return redisCli.Del(channel).Err() != nil
	}

	key, ok := r.mapChanKey.Load(channel)
	if ok {
		r.mapChanKey.Delete(channel)
		r.mapKeyChan.Delete(key)
		return true
	}
	return false
}

func (r *RoomKeysType) DeleteKey(key string) bool {
	if !saveInFile {
		return redisCli.Del(key).Err() != nil
	}

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
