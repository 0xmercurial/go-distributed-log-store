package logpack

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

/*
Log from scratch
-----------------------
- Log is composed of
  + Segments are composed of
    -> store files contain
	   + records
	-> log files contain
	   + record indexes
*/

var (
	/*
		- for defining the order in which bytes are written
		- little endianness is superior when you want flexibility in the size of the data being represented
	*/
	enc = binary.BigEndian
)

const (
	lenWidth = 8
)

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f), // writer will always write to specified file
	}, nil
}

//Append writes value bytes as a record to the file using the binary Write method
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pos = s.size
	//Writing length of record for later read at position
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}
	// log.Println("Record len: ", uint64(len(p)))
	// log.Println("Appending: ", p)
	// log.Println("Pos: ", pos)

	//Writing actual record
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}

	w += lenWidth
	s.size += uint64(w)
	return uint64(w), pos, nil
}

// Read reads bytes at a given postion and returns them
func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil { //ensure no data is still in the buffer
		return nil, err
	}
	//Read the first 8 bytes for length of record at position
	size := make([]byte, lenWidth)
	_, err := s.File.ReadAt(size, int64(pos))
	if err != nil {
		return nil, err
	}
	// log.Println("N-read: ", n)
	// log.Println("Size: ", size) //remove logs after test
	// log.Println("Sint: ", enc.Uint64(size))

	//Read the next `size` number of bytes after the length record.
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil { // try and read record of length lenWidth starting at position pos
		return nil, err
	}
	// log.Println("Read: ", b) //remove logs after test
	return b, nil
}

//TODO: This seems redudant?
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

//Close persists data before closing the store file
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}
