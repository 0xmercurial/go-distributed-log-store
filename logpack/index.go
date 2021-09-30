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

type Config struct {
	Segment struct {
		MaxStoreBytes uint64
		MaxIndexBytes uint64
		InitialOffset uint64
	}
}

func newIndex(f *os.File, c Config) (*index, error) {

	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	//TODO: Implement config struct
	size := uint64(fi.Size())
	err = os.Truncate(
		f.Name(), int64(c.Segment.MaxIndexBytes),
	)
	if err != nil {
		return nil, err
	}
	//TODO: Research beter configs for indexing
	idx := &index{
		file: f,
		size: size,
	}
	mmap, err := gommap.Map(
		idx.file.Fd(),
		gommap.PROT_READ|gommap.PROT_WRITE,
		gommap.MAP_SHARED,
	)
	if err != nil {
		return nil, err
	}
	idx.mmap = mmap
	return idx, nil
}

func (i *index) Close() error {
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}
	if err := i.file.Sync(); err != nil {
		return err
	}
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close()
}
