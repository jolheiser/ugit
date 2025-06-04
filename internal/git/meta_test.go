package git

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestTagSet(t *testing.T) {
	set := make(TagSet)
	assert.Equal(t, 0, len(set))
	assert.Equal(t, 0, len(set.Slice()))

	set.Add("foo")
	assert.Equal(t, 1, len(set))
	assert.Equal(t, 1, len(set.Slice()))
	assert.True(t, set.Contains("foo"))

	set.Add("bar")
	assert.Equal(t, 2, len(set))
	assert.Equal(t, 2, len(set.Slice()))
	assert.True(t, set.Contains("foo"))
	assert.True(t, set.Contains("bar"))

	set.Add("bar")
	assert.Equal(t, 2, len(set))
	assert.Equal(t, 2, len(set.Slice()))
	assert.True(t, set.Contains("foo"))
	assert.True(t, set.Contains("bar"))

	set.Remove("foo")
	assert.Equal(t, 1, len(set))
	assert.Equal(t, 1, len(set.Slice()))
	assert.False(t, set.Contains("foo"))
	assert.True(t, set.Contains("bar"))

	set.Add("foo")
	set.Add("baz")
	j, err := json.Marshal(set)
	assert.NoError(t, err)
	assert.Equal(t, `["bar","baz","foo"]`, string(j))

	set = make(TagSet)
	b := []byte(`["foo","bar","baz"]`)
	err = json.Unmarshal(b, &set)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(set))
	assert.Equal(t, 3, len(set.Slice()))
	assert.True(t, set.Contains("foo"))
	assert.True(t, set.Contains("bar"))
	assert.True(t, set.Contains("baz"))
}
