package logcomponents

import (
	"io/ioutil"
	prolog "logstore/internal/log/proto"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

var dir string
var err error

func newTestLog() (*Log, error) {
	dir, err = ioutil.TempDir("", "log-test")
	if err != nil {
		os.Exit(1)
	}
	c := Config{}
	c.Segment.MaxStoreBytes = 32
	log, err := NewLog(dir, c)
	return log, err
}

func TestAppendRead(t *testing.T) {
	log, err := newTestLog()
	assert.NoError(t, err)
	record := &prolog.Record{
		Value: []byte("record"),
	}
	off, err := log.Append(record)
	if err != nil {
		t.Error(err)
	}
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), off)

	read, err := log.Read(off)
	assert.NoError(t, err)
	assert.Equal(t, record.Value, read.Value)
}

func TestOutOfRangeErr(t *testing.T) {
	log, err := newTestLog()
	defer os.RemoveAll(dir)
	assert.NoError(t, err)
	read, err := log.Read(1) // should fail w/ fresh log
	assert.Nil(t, read)
	protoErr := err.(prolog.ErrOffOutOfRange) //asserting type of error
	assert.Equal(t, uint64(1), protoErr.Offset)
}

func TestInitExisting(t *testing.T) {
	log, err := newTestLog()
	defer os.RemoveAll(dir)
	assert.NoError(t, err)
	record := &prolog.Record{
		Value: []byte("record"),
	}
	for i := 0; i < 3; i++ {
		_, err := log.Append(record)
		assert.NoError(t, err)
	}
	assert.NoError(t, log.Close())

	//Initial log
	off, err := log.LowestOffset()
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), off)
	off, err = log.HighestOffset()
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), off)

	//New log initiated from existing
	new, err := NewLog(log.Dir, log.Config)

	off, err = new.LowestOffset()
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), off)
	off, err = new.HighestOffset()
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), off)
}

func TestReader(t *testing.T) {
	log, err := newTestLog()
	defer os.RemoveAll(dir)
	assert.NoError(t, err)
	record := &prolog.Record{
		Value: []byte("record"),
	}
	off, err := log.Append(record)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), off)

	rdr := log.Reader()
	byt, err := ioutil.ReadAll(rdr)
	assert.NoError(t, err)

	readRec := &prolog.Record{}
	err = proto.Unmarshal(byt[lenWidth:], readRec)
	assert.Equal(t, record.Value, readRec.Value)
}

func TestTruncate(t *testing.T) {
	log, err := newTestLog()
	defer os.RemoveAll(dir)
	assert.NoError(t, err)
	record := &prolog.Record{
		Value: []byte("record"),
	}
	for i := 0; i < 3; i++ {
		_, err := log.Append(record)
		assert.NoError(t, err)
	}

	err = log.Truncate(1)
	assert.NoError(t, err)

	_, err = log.Read(0)
	assert.Error(t, err)
}
