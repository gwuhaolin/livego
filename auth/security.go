package auth

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-redis/redis/v7"
	log "github.com/sirupsen/logrus"

	cli "github.com/gwuhaolin/livego/auth/entity/client"
	sess "github.com/gwuhaolin/livego/auth/entity/session"
	"github.com/gwuhaolin/livego/configure"
)

var Auth *Security

type Security struct {
	redisCli *redis.Client
	Allow    func(string, string) (bool, error)
}

func init() {

	protected := configure.Config.GetBool("is_protected")
	log.Infof("security.init(): %v", protected)

	// Identify if Authentication is enabled:
	Auth = New(protected)
	// =====================================================

	log.Infof("redis_addr: %s | redis_pwd: %s | redis_db: %d",
		configure.Config.GetString("redis_addr"),
		configure.Config.GetString("redis_pwd"),
		configure.Config.GetInt("redis_db"))
	if len(configure.Config.GetString("redis_addr")) != 0 {

		Auth.redisCli = redis.NewClient(&redis.Options{
			Addr:     configure.Config.GetString("redis_addr"),
			Password: configure.Config.GetString("redis_pwd"),
			DB:       configure.Config.GetInt("redis_db"),
		})
		_, err := Auth.redisCli.Ping().Result()
		if err != nil {
			log.Panic("Redis: ", err)
		}
		log.Info("Redis connected")

	}

}

func New(protected bool) (security *Security) {

	security = &Security{}
	if protected {
		security.Allow = security.allow
	} else {
		security.Allow = func(string, string) (bool, error) {
			log.Debugf("security.Allow(): default true")
			return true, nil // Default true
		}
	}
	return

}

func (security *Security) allow(ipAddr, value string) (allowed bool, err error) {

	allowed = false
	if url, err := url.Parse(value); err == nil {
		user := url.Query().Get("u")
		key := url.Query().Get("k")
		log.Debugf("security.Allow(): [%s]-[%s]", user, len(key) > 0)

		if client, err := security.findClient(user); err == nil {
			if client != nil && client.Active {

				lsk := client.LastSessionKey

				session, err := security.findSession(lsk)
				if err == redis.Nil || session == nil {
					session = sess.New(user, ipAddr)
					if err = security.saveSession(session); err == nil {
						client.LastSessionKey = session.Key()
						if err = security.saveClient(client); err == nil {
							log.Debugf("security.Allow(): %+v", session)
							allowed = true
						}
					}
				} else {
					// Validate the IP Address from source to identify if it is another session to the same user:
					allowed = (strings.Split(ipAddr, ":")[0] == strings.Split(session.IpAddress, ":")[0])
				}
			}
		}
	}
	if err != nil {
		log.Errorf("security.Allow(): %s", err.Error())
	}
	return

}

func (s *Security) findClient(code string) (*cli.Client, error) {

	if value, err := s.redisCli.Get(fmt.Sprintf("client-%s", code)).Result(); err == nil {
		return cli.FromJson([]byte(value))
	} else {
		return nil, err
	}

}

func (s *Security) saveClient(client *cli.Client) (err error) {

	content, _ := client.ToJson()
	cmd_val, err := s.redisCli.Set(fmt.Sprintf("client-%s", client.Code), string(content), 0).Result()
	log.Debugf("security.saveClient(): %s", cmd_val)
	return

}

func (s *Security) findSession(key string) (*sess.Session, error) {

	if value, err := s.redisCli.Get(key).Result(); err == nil {
		return sess.FromJson([]byte(value))
	} else {
		return nil, err
	}

}

func (s *Security) saveSession(session *sess.Session) (err error) {

	content, _ := session.ToJson()
	cmd_val, err := s.redisCli.Set(session.Key(), string(content), 0).Result()
	log.Debugf("security.saveSession(): %s", cmd_val)
	return

}
