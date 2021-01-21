package plugin

import "testing"

func TestExtractPathParams(t *testing.T) {
	g := New()

	p1 := "/{hello}/{world}"
	paths := g.extractPathParams(p1)
	t.Log(paths)
	p2 := "/{dsf/asdf"
	paths = g.extractPathParams(p2)
	t.Log(paths)
}
