package logcomponents

import (
	"sync"
)

type Log struct {
	mu sync.RWMutex

	Dir    string
	Config Config

	activeSegment *segment
	segments      []*segment
}

func NewLog(dir string, c Config) (*Log, error) {

	if c.Segment.MaxStoreBytes == 0 {
		c.Segment.MaxStoreBytes = 1024
	}

	if c.Segment.MaxIndexBytes == 0 {
		c.Segment.MaxIndexBytes = 1024
	}
	l := &Log{
		Dir:    dir,
		Config: c,
	}

	return l, l.setup()
}

func (l *Log) setup() error {
	return nil
	// files, err := ioutil.ReadDir(l.Dir)
	// if err != nil {
	// 	return err
	// }
	// var baseOffsets []uint64
	// for _, file := range files {
	// 	offStr := strings.TrimSuffix(file.Name(), path.Ext(file.Name()))
	// }
}
