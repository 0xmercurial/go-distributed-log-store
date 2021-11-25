package logcomponents

import (
	"io"
	"io/ioutil"
	prolog "logstore/internal/log/proto"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSegment(t *testing.T) {
	//Test that append/read method for segment works.
	dir, _ := ioutil.TempDir("", "seg-test")
	defer os.Remove(dir)

	want := &prolog.Record{Value: []byte("hello world")}
	t.Log(want)

	c := Config{}
	c.Segment.MaxStoreBytes = 1024
	c.Segment.MaxIndexBytes = entWidth * 3

	s, err := newSegment(dir, 16, c)
	assert.Equal(t, err, nil)
	assert.Equal(t, uint64(16), s.nextOffset)
	assert.Equal(t, false, s.IsMaxed())

	for i := uint64(0); i < 3; i++ {
		off, err := s.Append(want)
		assert.Equal(t, err, nil)
		assert.Equal(t, 16+i, off)

		got, err := s.Read(off)
		assert.NoError(t, err)
		assert.Equal(t, want.Value, got.Value)
	}

	_, err = s.Append(want)
	assert.Equal(t, io.EOF, err)
	assert.True(t, s.IsMaxed())

	c.Segment.MaxStoreBytes = uint64(len(want.Value) * 3)
	c.Segment.MaxIndexBytes = 1024

	s, err = newSegment(dir, 16, c)
	assert.NoError(t, err)
	assert.True(t, s.IsMaxed())

	err = s.Remove()
	assert.NoError(t, err)
	s, err = newSegment(dir, 16, c)
	assert.NoError(t, err)
	assert.False(t, s.IsMaxed())

}
