package plugin

import "testing"

func TestLinkComponents_Range(t *testing.T) {
	l := NewLinkComponents()
	c1 := &Component{Name: "1"}
	c2 := &Component{Name: "2"}
	l.Push(c1)
	l.Push(c2)
	ch := make(chan struct{}, 1)
	go func() {
		l.Range(func(c *Component) {
			t.Logf(c.Name)
		})
		ch <- struct{}{}
	}()
	l.Push(&Component{Name: "3"})
	l.Push(&Component{Name: "4"})
	l.Push(&Component{Name: "5"})
	l.Push(&Component{Name: "6"})
	l.Push(&Component{Name: "6"})
	l.Push(&Component{Name: "7"})
	<-ch
}
