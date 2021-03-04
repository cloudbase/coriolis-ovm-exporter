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
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

type reflinkArgs struct {
	OldPath  uint64
	NewPath  uint64
	Preserve uint64
}

const (
	// OCFS2IOCReflink is the OCFS2 ioctl for creating a cow file
	OCFS2IOCReflink = 1075343108
)

// IOctlOCFS2Reflink creates a reflinked copy (copy-on-write) on an OCFS2
func IOctlOCFS2Reflink(src, dst string) error {
	srcCopy := fmt.Sprintf(src)
	dstCopy := fmt.Sprintf(dst)
	params := reflinkArgs{
		OldPath:  *(*uint64)(unsafe.Pointer(&srcCopy)),
		NewPath:  *(*uint64)(unsafe.Pointer(&dstCopy)),
		Preserve: 1,
	}

	fd, err := os.Open(srcCopy)
	if err != nil {
		return errors.Wrap(err, "opening file")
	}

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd.Fd(), OCFS2IOCReflink, uintptr(unsafe.Pointer(&params))); err != 0 {
		return errors.Wrap(err, "running ioctl")
	}
	return nil
}
