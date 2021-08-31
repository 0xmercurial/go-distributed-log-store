package log

import (
	"fmt"
	"sync"
)

//Struct for unmarshalling JSON bytes into.
type Record struct {
	Value  []byte `json"value"`
	Offset uint64 `json"offset"` // Offset will only ever be positive integer. Only set by Append().
}

var ErrOffsetNotFound = fmt.Errorf("offset not found")

type Log struct {
	mu      sync.Mutex
	records []Record
}

func NewLog() *Log {
	return &Log{}
}

func (l *Log) Append(record Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	record.Offset = uint64(len(l.records))
	l.records = append(l.records, record)
	return record.Offset, nil
}

func (l *Log) Read(offset uint64) (Record, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if offset >= uint64(len(l.records)) {
		return Record{}, ErrOffsetNotFound
	}
	return l.records[offset], nil
}

func (l *Log) ReadRange(offset1 uint64, offset2 uint64) ([]Record, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if offset1 >= uint64(len(l.records)) {
		return []Record{}, ErrOffsetNotFound
	}
	return l.records[offset1:MIN(offset2, uint64(len(l.records)))], nil
}

func (l *Log) ReadAll() []Record {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.records
}

func MIN(a uint64, b uint64) uint64 {
	if a <= b {
		return a
	}
	return b
}
