package name

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	n := New()
	assert.NotEqual(t, n, nil)
}

func Test_name_Digit(t *testing.T) {
	n := New().Digit(1)
	assert.Equal(t, n.b.Len(), 1)
	n.Digit(2)
	assert.Equal(t, n.b.Len(), 3)
	n.Digit(3)
	assert.Equal(t, n.b.Len(), 6)

	t.Log(n.String())
}

func Test_name_Lower(t *testing.T) {
	n := New().Lower(1)
	assert.Equal(t, n.b.Len(), 1)

	n1 := New().Lower(1)

	assert.NotEqual(t, n.String(), n1.String())

	n.Lower(3)
	assert.Equal(t, n.b.Len(), 4)

	t.Log(n.String())
}

func Test_name_Letter(t *testing.T) {
	n := New().Letter(5)
	assert.Equal(t, n.b.Len(), 5)

	t.Log(n.String())
}

func Test_name_Upper(t *testing.T) {
	n := New().Upper(1)
	assert.Equal(t, n.b.Len(), 1)

	n1 := New().Upper(1)

	assert.NotEqual(t, n.String(), n1.String())

	n.Upper(3)
	assert.Equal(t, n.b.Len(), 4)

	t.Log(n.String())
}

func Test_name_Any(t *testing.T) {
	n := New().Any(5)
	assert.Equal(t, n.b.Len(), 5)

	t.Log(n.String())
}