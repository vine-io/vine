package plugin

import (
	"container/list"
	"sync"
)

type LinkComponents struct {
	sync.RWMutex
	data *list.List
	m    map[string]struct{}
}

func NewLinkComponents() *LinkComponents {
	return &LinkComponents{
		data: list.New(),
		m:    map[string]struct{}{},
	}
}

func (l *LinkComponents) Push(c *Component) {
	l.Lock()
	defer l.Unlock()
	if _, ok := l.m[c.Name]; ok {
		return
	}
	l.data.PushBack(c)
	l.m[c.Name] = struct{}{}
}

func (l *LinkComponents) Range(fn func(*Component)) {
	l.RLock()
	ptr := l.data.Front()
	l.RUnlock()
	for {
		if ptr == nil {
			break
		}
		fn(ptr.Value.(*Component))
		l.RLock()
		ptr = ptr.Next()
		l.RUnlock()
	}
	return
}
