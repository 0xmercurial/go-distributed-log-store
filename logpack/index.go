package logpack

import (
	"os"

	"github.com/tysontate/gommap"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

func newIndex(f *os.File) (*index, error) {

	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil , err
	}

	//TODO: Implement config struct
	size := uint64(fi.Size())
	if err = os.Truncate(
		f.Name(), int64()
	)

	idx := &index {
		file: f,
	}
}