package logcomponents

import (
	"fmt"
	"io/ioutil"
	"logpack/internal/log/proto"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
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

/*
setup intitalizes the log and inits log primitives.
If there are existing segments to intialize state, otherwise starts
with new segments
*/
func (l *Log) setup() error {
	files, err := ioutil.ReadDir(l.Dir)
	if err != nil {
		return err
	}
	var baseOffsets []uint64
	for _, file := range files {
		offStr := strings.TrimSuffix(
			file.Name(),
			path.Ext(file.Name()),
		)
		off, _ := strconv.ParseUint(offStr, 10, 0)
		baseOffsets = append(baseOffsets, off)
	}
	sort.Slice(baseOffsets, func(i, j int) bool {
		return baseOffsets[i] < baseOffsets[j]
	})

	for i := 0; i < len(baseOffsets); i++ {
		if err = l.newSegment(
			l.Config.Segment.InitialOffset,
		); err != nil {
			return err
		}
	}
	return nil
}

/*
	Append appends a record to the log, specifically to the active segment.
*/
func (l *Log) Append(record *proto.Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	off, err := l.activeSegment.Append(record)
	if err != nil {
		return 0, err
	}

	if l.activeSegment.IsMaxed() {
		err != l.newSegment(off+1)
	}

	return off, err
}

/*
 */
func (l *Log) Read(off uint64) (*proto.Record, error) {
	/*
		A read/write mutex allows all the readers to access mem at the same time,
		but a writer will lock out everyone else.
	*/
	l.mu.RLock()
	defer l.mu.RUnlock() // readers holding lock only have to wait to writers

	var s *segment
	for _, segment := range l.segments {
		if segment.baseOffset <= off && off < segment.nextOffset {
			s = segment
			break
		}
	}
	if s == nil || s.nextOffset <= off {
		return nil, fmt.Errorf("offset of of range: %d", off)
	}
	return s.Read(off)
}

/*
Close
*/
func (l *Log) Close() error {
	l.mu.Lock() //ensure no more reads/writes occur
	defer l.mu.Unlock()

	for _, segment := range l.segments {
		if err := segment.Close(); err != nil { //close out existing segments
			return err
		}
	}
	return nil
}

/*
Remove
*/
func (l *Log) Remove() error {
	if err := l.Close(); err != nil {
		return err
	}
	return os.RemoveAll(l.Dir)
}

/*
Reset
*/
func (l *Log) Reset() error {
	if err := l.Remove(); err != nil {
		return err
	}
	return l.setup()
}
