package replication

import (
	"key-value/lib/routers"
)

type Client interface {
	AddNode(addr string)
	HandleRemoved(key string, version int64)
	HandleUpdated(key string, val string, version int64)
}

type client struct {
	nodes map[string]routers.Client
}

func NewClient() Client {
	return &client{make(map[string]routers.Client)}
}

func (c *client) AddNode(addr string) {
	c.nodes[addr] = nil
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
