package logpack

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	record = []byte("hello world")
	width  = uint64(len(record)) + lenWidth
)

func TestStoreAppendRead(t *testing.T) {
	f, err := ioutil.TempFile("", "store_append_read_test")
	assert.Equal(t, err, nil)
	defer os.Remove(f.Name())

	s, err := newStore(f)
	assert.Equal(t, err, nil)

	// t.Log("Record: ", record)
	testAppend(t, s)
	testRead(t, s)
	testReadAt(t, s)
}

func openFile(name string) (file *os.File, size int64, err error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, 0, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	return f, fi.Size(), nil
}

func testAppend(t *testing.T, s *store) {
	t.Helper()
	for i := uint64(1); i < 4; i++ {
		n, pos, err := s.Append(record)
		assert.Equal(t, err, nil)
		assert.Equal(t, pos+n, width*i)
	}
}

func testRead(t *testing.T, s *store) {
	t.Helper()
	var pos uint64
	for i := uint64(1); i < 4; i++ {
		read, err := s.Read(pos)
		assert.Equal(t, err, nil)
		assert.Equal(t, record, read)
		pos += width
	}
}

func testReadAt(t *testing.T, s *store) {
	t.Helper()
	for i, off := uint64(1), int64(0); i < 4; i++ {
		b := make([]byte, lenWidth)
		n, err := s.ReadAt(b, off)
		assert.Equal(t, err, nil)
		assert.Equal(t, lenWidth, n)
		off += int64(n)

		size := enc.Uint64(b)
		b = make([]byte, size)
		n, err = s.ReadAt(b, off)
		assert.Equal(t, err, nil)
		assert.Equal(t, record, b)
		assert.Equal(t, int(size), n)
		off += int64(n)
	}
}

func TestStoreClose(t *testing.T) {
	f, err := ioutil.TempFile("", "store_close_test")
	assert.Equal(t, err, nil)
	defer os.Remove(f.Name())

	s, err := newStore(f)
	assert.Equal(t, err, nil)
	_, _, err = s.Append(record)
	assert.Equal(t, err, nil)
	assert.Equal(t, err, nil)

	f, before, err := openFile(f.Name())
	assert.Equal(t, err, nil)

	err = s.Close()
	assert.Equal(t, err, nil)

	_, after, err := openFile(f.Name())
	assert.Equal(t, err, nil)
	assert.Equal(t, after > before, true)
}
