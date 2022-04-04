package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

type cacheItem struct {
	key   Key
	value interface{}
}

func NewCache(capacity int) Cache {
	if capacity < 1 {
		capacity = 1
	}
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	newCacheValue := cacheItem{
		key:   key,
		value: value,
	}
	element, exist := c.items[key]
	if exist {
		element.Value = newCacheValue
		c.queue.MoveToFront(element)
		return true
	} else {
		if c.queue.Len() >= c.capacity {
			cacheElement := c.queue.Back().Value.(cacheItem)
			delete(c.items, cacheElement.key)
			c.queue.Remove(c.queue.Back())
		}
		c.queue.PushFront(newCacheValue)
		c.items[key] = c.queue.Front()
		return false
	}
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	element, exist := c.items[key]
	if !exist {
		return nil, false
	}
	c.queue.MoveToFront(element)
	return element.Value.(cacheItem).value, true
}

func (c *lruCache) Clear() {
	c.items = make(map[Key]*ListItem, c.capacity)
	c.queue = NewList()
}
