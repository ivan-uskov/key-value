package replication

import (
	"key-value/lib/routers"
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type Client interface {
	HandleNewNodesRequest(r routers.Request) (string, error)
	HandleRegisterRequest(r routers.Request) (string, error)
	HandleRemoved(key string, version int64)
	HandleUpdated(key string, val string, version int64)
}

type client struct {
	nodes       map[string]routers.Client
	selfAddress string
}

func NewClient(selfAddress string) Client {
	return &client{make(map[string]routers.Client), selfAddress}
}

func (c *client) HandleRegisterRequest(r routers.Request) (string, error) {
	if c.selfAddress != r.Option1 {
		log.WithField(`addr`, r.Option1).Info(`new node registered`)
		c.nodes[r.Option1] = nil
	}
	return ``, nil
}

func (c *client) HandleNewNodesRequest(r routers.Request) (string, error) {
	var addrs []string
	err := json.Unmarshal([]byte(r.Option1), &addrs)
	if err != nil {
		return ``, err
	}

	for _, v := range addrs {
		if c.selfAddress != v {
			c.nodes[v] = nil
		}
	}

	log.WithFields(log.Fields{`nodes`: c.nodes}).Info(`got new nodes`)
	go func() {
		c.forNodes(func(con routers.Client) {
			c.sync(con, routers.Request{
				Action:  register,
				Option1: c.selfAddress,
			})
		})
	}()

	return ``, nil
}

func (c *client) HandleRemoved(key string, version int64) {
	log.WithFields(log.Fields{`key`: key, `ver`: version}).Info(`sync remove`)
	c.forNodes(func(con routers.Client) {
		c.sync(con, routers.Request{
			Action:  removed,
			Option1: key,
			Version: version,
		})
	})
}

func (c *client) HandleUpdated(key string, val string, version int64) {
	log.WithFields(log.Fields{`key`: key, `val`: val, `ver`: version}).Info(`sync update`)
	c.forNodes(func(con routers.Client) {
		c.sync(con, routers.Request{
			Action:  updated,
			Option1: key,
			Option2: val,
			Version: version,
		})
	})
}

func (c *client) sync(con routers.Client, r routers.Request) {
	log.WithFields(log.Fields{`r`: r}).Info(`sync`)
	resp, err := con.SendSync(r)
	if err != nil {
		log.Error(err)
	} else {
		log.WithFields(log.Fields{`resp`: resp}).Info(`sync sent`)
	}
}

func (c *client) forNodes(handler func(routers.Client)) {
	for addr, con := range c.nodes {
		if con == nil {
			con, err := routers.NewClient(addr, path)
			if err != nil {
				log.WithFields(log.Fields{`addr`: addr}).Error(err)
				continue
			}
			c.nodes[addr] = con
		}

		handler(c.nodes[addr])
	}
}
