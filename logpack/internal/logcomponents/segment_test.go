package logcomponents

import (
	"io/ioutil"
	prolog "logpack/internal/log/proto"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSegment(t *testing.T) {
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
	}

}
