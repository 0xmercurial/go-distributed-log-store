package logcomponents

import (
	"io/ioutil"
	"logstore/internal/log/proto"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func helper() (string, Config) {
	dir, err := ioutil.TempDir("", "log-test")
	if err != nil {
		os.Exit(1)
	}
	c := Config{}
	c.Segment.MaxStoreBytes = 32
	return dir, c
}

// func TestLog(t *testing.T) {
// 	for scenario, fn := range map[string]func(
// 		t *testing.T, log *Log,
// 	){
// 		"append"
// 	}
// }

func TestAppendRead(t *testing.T) {
	dir, c := helper()
	log, err := NewLog(dir, c)
	assert.NoError(t, err)
	newRec := &proto.Record{
		Value: []byte("hello world"),
	}
	off, err := log.Append(newRec)
	if err != nil {
		t.Error(err)
	}
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), off)

	read, err := log.Read(off)
	assert.NoError(t, err)
	assert.Equal(t, newRec.Value, read.Value)
}
