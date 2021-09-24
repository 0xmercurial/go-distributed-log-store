package logpack

import (
	"encoding/binary"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	record = []byte("henlo world")
	width  = uint64(len(record)) + lenWidth
)

func TestEndian(t *testing.T) {
	buf := make([]byte, 16)
	enc.PutUint16(buf, uint16(50))
	t.Log(enc.Uint16(buf))
	t.Log(binary.LittleEndian.Uint16(buf))
}

func TestStoreAppendRead(t *testing.T) {
	f, err := ioutil.TempFile("", "store_append_read_test")
	assert.Equal(t, err, nil)
	defer os.Remove(f.Name())

	s, err := newStore(f)
	assert.Equal(t, err, nil)

	testAppend(t, s)
	testRead(t, s)
	testReadAt(t, s)
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
}
