package third_party

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStatic(t *testing.T) {
	fs := GetStatic()
	assert.NotEmpty(t, fs, "get static can't be empty")
}
