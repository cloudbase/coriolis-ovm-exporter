package internal

import (
	"fmt"

	"github.com/pkg/errors"
)

func getFileExtents(filePath string) ([]Chunk, error) {
	extents, err := GetExtents(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "fetching fiemap")
	}

	if len(extents) == 0 {
		return nil, fmt.Errorf("failed to get extents")
	}

	ret := make([]Chunk, len(extents))

	for idx, ext := range extents {
		ret[idx] = Chunk{
			Length:   ext.Length,
			Start:    ext.Logical,
			Physical: ext.Physical,
		}
	}
	return ret, nil
}

func squashChunks(chunks []Chunk) []Chunk {
	if chunks == nil || len(chunks) == 0 {
		return []Chunk{}
	}

	tmpLogical := chunks[0].Start
	tmpPhysical := chunks[0].Physical
	tmpLength := chunks[0].Length

	var squashed []Chunk

	for i := 1; i < len(chunks); i++ {
		if (tmpLogical + tmpLength) == chunks[i].Start {
			tmpLength += chunks[i].Length
			continue
		}

		squashed = append(squashed, Chunk{
			Start:    tmpLogical,
			Length:   tmpLength,
			Physical: tmpPhysical,
		})
		tmpLogical = chunks[i].Start
		tmpPhysical = chunks[i].Physical
		tmpLength = chunks[i].Length
	}

	squashed = append(squashed, Chunk{
		Start:    tmpLogical,
		Length:   tmpLength,
		Physical: tmpPhysical,
	})

	return squashed
}

func getSquashedFileExtents(filePath string) ([]Chunk, error) {
	chunks, err := getFileExtents(filePath)

	if err != nil {
		return nil, err
	}

	return squashChunks(chunks), nil
}
