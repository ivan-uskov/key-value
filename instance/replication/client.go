package replication

import (
	"key-value/lib/routers"
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type Client interface {
	HandleNewNodesRequest(r routers.Request) (string, error)
	HandleRemoved(key string, version int64)
	HandleUpdated(key string, val string, version int64)
}

type client struct {
	nodes map[string]routers.Client
	selfAddress string
}

func NewClient(selfAddress string) Client {
	return &client{make(map[string]routers.Client), selfAddress}
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

	return ``, nil
}

func (c *client) HandleRemoved(key string, version int64) {
	c.forNodes(func(con routers.Client) {
		con.SendSync(routers.Request{
			Action:  removed,
			Option1: key,
			Version: version,
		})
	})
}

func (c *client) HandleUpdated(key string, val string, version int64) {
	c.forNodes(func(con routers.Client) {
		con.SendSync(routers.Request{
			Action:  updated,
			Option1: key,
			Option2: val,
			Version: version,
		})
	})
}

func (c *client) forNodes(handler func(routers.Client)) {
	for addr, con := range c.nodes {
		if con == nil {
			con, err := routers.NewClient(addr, `/ws`)
			if err != nil {
				continue
			}
			c.nodes[addr] = con
		}

		handler(con)
	}
}
