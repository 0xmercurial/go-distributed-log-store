package logcomponents

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndex(t *testing.T) {
	//Building index from new/temp file
	f, err := ioutil.TempFile(os.TempDir(), "index_test")
	assert.Equal(t, err, nil)
	defer os.Remove(f.Name())

	c := Config{}
	c.Segment.MaxIndexBytes = 1024
	idx, err := newIndex(f, c)
	assert.Equal(t, err, nil)

	//Reading index w/ empty mmap
	_, _, err = idx.Read(-1) // <- any integer value will trigger an error at this stage
	assert.Error(t, err, nil)
	assert.Equal(t, f.Name(), idx.Name())

	entries := []struct {
		Off uint32
		Pos uint64
	}{
		{Off: 0, Pos: 0},
		{Off: 1, Pos: 10},
	}

	for _, want := range entries {
		err = idx.Write(want.Off, want.Pos)
		assert.Equal(t, err, nil)
		_, pos, err := idx.Read(int64(want.Off))
		assert.Equal(t, err, nil)
		assert.Equal(t, want.Pos, pos)
	}

	_, _, err = idx.Read(int64(len(entries)))
	assert.Equal(t, io.EOF, err)
	idx.Close()

	// BUilding index from existing file
	f, _ = os.OpenFile(f.Name(), os.O_RDWR, 6000)
	idx, err = newIndex(f, c)
	assert.Equal(t, err, nil)
	off, pos, err := idx.Read(-1)
	assert.Equal(t, err, nil)
	assert.Equal(t, uint32(1), off)
	assert.Equal(t, entries[1].Pos, pos)

}
