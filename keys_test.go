package keys

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssign(t *testing.T) {
	s := new(Shortcuts)
	s.AutoAssign("a", "g1")
	s.AutoAssign("b", "g1")
	s.AutoAssign("c", "g2")
	s.AutoAssign("d", "g2")

	assert.True(t, s.index("a", "g1") > -1)
	assert.True(t, s.index("b", "g1") > -1)
	assert.True(t, s.index("c", "g1") == -1)
	assert.True(t, s.index("d", "g1") == -1)

	assert.True(t, s.index("a", "g2") == -1)
	assert.True(t, s.index("b", "g2") == -1)
	assert.True(t, s.index("c", "g2") > -1)
	assert.True(t, s.index("d", "g2") > -1)

	assert.True(t, s.index("a", "") == -1)
	assert.True(t, s.index("b", "") == -1)
	assert.True(t, s.index("c", "") == -1)
	assert.True(t, s.index("d", "") == -1)

	// a exists in group g1.
	a, has := s.MapsTo('a', "g1")
	assert.True(t, has)
	assert.Equal(t, "a", a)

	// a doesn't exist in group g2.
	a, has = s.MapsTo('a', "g2")
	assert.False(t, has)

	// a doesn't exist in empty group.
	a, has = s.MapsTo('a', "")
	assert.False(t, has)

	// Can't assign the same shortcut in the same group.
	s.BestEffortAssign('a', "e", "g1")
	e, has := s.MapsTo('e', "g1")
	assert.True(t, has)
	assert.Equal(t, "e", e)

	// Reuse the same shortcut in different group.
	s.BestEffortAssign('a', "f", "g2")
	f, has := s.MapsTo('a', "g2")
	assert.True(t, has)
	assert.Equal(t, "f", f)

	s.Print("")
	s.Print("g1")
	s.Print("g2")
	s.Validate()

	tf, err := ioutil.TempFile("", "keys")
	assert.NoError(t, err)
	s.Persist(tf.Name())

	ds := ParseConfig(tf.Name())
	for i, k := range s.Keys {
		assert.Equal(t, k, ds.Keys[i])
	}
}
