package cache

import (
	"github.com/mogaika/god_of_war_browser/pack/wad"
)

type Cache struct {
	d map[wad.TagId]interface{}
}

func (c *Cache) Add(id wad.TagId, d interface{}) {
	c.d[id] = d
}

func (c *Cache) Get(id wad.TagId) interface{} {
	if d, e := c.d[id]; e {
		return d
	} else {
		return nil
	}
}

func NewCache() *Cache {
	return &Cache{d: make(map[wad.TagId]interface{})}
}
