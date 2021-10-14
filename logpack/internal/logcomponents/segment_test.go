package logcomponents

import (
	"io/ioutil"
	prolog "logpack/internal/log/proto"
	"os"
	"testing"
)

func TestSegment(*testing.T) {
	dir, _ := ioutil.TempDir("", "seg-test")
	defer os.Remove(dir)

	want := &prolog.Record{Value: []byte("hello world")}

	c := Config{}
	c.Segment.MaxStoreBytes = 1024
	c.Segment.MaxIndexBytes = entWidth * 3
}
