package index_test

import (
	"testing"

	"github.com/lack-io/vine/service/ent/schema/index"

	"github.com/stretchr/testify/require"
)

func TestIndex(t *testing.T) {
	idx := index.Fields("name", "address").
		StorageKey("user_index").
		Descriptor()
	require.Empty(t, idx.Edges)
	require.False(t, idx.Unique)
	require.Equal(t, "user_index", idx.StorageKey)
	require.Equal(t, []string{"name", "address"}, idx.Fields)

	idx = index.Fields("name", "address").
		Unique().
		Descriptor()
	require.Empty(t, idx.Edges)
	require.True(t, idx.Unique)
	require.Empty(t, idx.StorageKey)
	require.Equal(t, []string{"name", "address"}, idx.Fields)

	idx = index.Fields("name", "address").
		Edges("parent", "type").
		Unique().
		Descriptor()
	require.Equal(t, []string{"parent", "type"}, idx.Edges)
	require.True(t, idx.Unique)
	require.Equal(t, []string{"name", "address"}, idx.Fields)
}
