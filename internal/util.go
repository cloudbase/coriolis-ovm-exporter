// Coriolis OVM exporter
// Copyright (C) 2021 Cloudbase Solutions SRL
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package internal

import (
	"github.com/pkg/errors"

	"coriolis-ovm-exporter/apiserver/params"
)

func getFileExtents(filePath string) ([]params.Chunk, error) {
	extents, err := GetExtents(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "fetching fiemap")
	}

	if len(extents) == 0 {
		// This is a thinly provisioned disk, with nothing written
		// inside it.
		return []params.Chunk{}, nil
	}

	ret := make([]params.Chunk, len(extents))

	for idx, ext := range extents {
		ret[idx] = params.Chunk{
			Length:   ext.Length,
			Start:    ext.Logical,
			Physical: ext.Physical,
		}
	}
	return ret, nil
}

// SquashChunks squashes continuous chunks into one chunk.
func SquashChunks(chunks []params.Chunk) []params.Chunk {
	if chunks == nil || len(chunks) == 0 {
		return []params.Chunk{}
	}

	tmpLogical := chunks[0].Start
	tmpPhysical := chunks[0].Physical
	tmpLength := chunks[0].Length

	var squashed []params.Chunk

	for i := 1; i < len(chunks); i++ {
		if (tmpLogical + tmpLength) == chunks[i].Start {
			tmpLength += chunks[i].Length
			continue
		}

		squashed = append(squashed, params.Chunk{
			Start:    tmpLogical,
			Length:   tmpLength,
			Physical: tmpPhysical,
		})
		tmpLogical = chunks[i].Start
		tmpPhysical = chunks[i].Physical
		tmpLength = chunks[i].Length
	}

	squashed = append(squashed, params.Chunk{
		Start:    tmpLogical,
		Length:   tmpLength,
		Physical: tmpPhysical,
	})

	return squashed
}

func getSquashedFileExtents(filePath string) ([]params.Chunk, error) {
	chunks, err := getFileExtents(filePath)

	if err != nil {
		return nil, err
	}

	return SquashChunks(chunks), nil
}
