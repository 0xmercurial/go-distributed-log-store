package logpack

import (
	"io"
	"os"

	"github.com/tysontate/gommap"
)

var (
	//position of entry in file is offset * entWidth
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap //map underlying file
	size uint64      //track size of file pointed to
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
		idx.file.Fd(), //memory map requires ptr to underlying file
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
	//Ensure mmap has synced data to file and that file has flushed contents to stable store.
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}
	if err := i.file.Sync(); err != nil {
		return err
	}
	//Remove blank space at end of file
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close()
}

func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	//Read the last entry
	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}
	pos = uint64(out) * entWidth
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}
	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return out, pos, nil
}

func (i *index) Write(off uint32, pos uint64) error {
	//Make sure there is space to write entry
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}
	// Put offset and position into memory map
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	//Increase size
	i.size += uint64(entWidth)
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}
