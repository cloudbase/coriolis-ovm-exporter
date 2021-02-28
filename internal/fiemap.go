package internal

import (
	"os"

	"github.com/pkg/errors"
	fibmap "github.com/rancher/go-fibmap"
)

const (
	// FeCount is the default extent count we request
	FeCount = 8000
)

func walkFiemap(fd fibmap.FibmapFile) ([]fibmap.Extent, error) {
	var ret []fibmap.Extent
	for {
		var last uint64 = 0
		if len(ret) > 0 {
			last = ret[len(ret)-1].Logical + ret[len(ret)-1].Length
		}
		extents, err := fibmap.Fiemap(fd.Fd(), last, fibmap.FIEMAP_MAX_OFFSET, FeCount)
		if int(err) != 0 {
			return nil, errors.Wrap(err, "fetching fiemap")
		}

		if len(extents) == 0 {
			break
		}
		ret = append(ret, extents...)
	}
	return ret, nil
}

// GetExtents returns a list of extents allocated to the file
func GetExtents(filename string) ([]fibmap.Extent, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "opening file")
	}
	defer fd.Close()
	fmFile := fibmap.NewFibmapFile(fd)

	return walkFiemap(fmFile)
}
