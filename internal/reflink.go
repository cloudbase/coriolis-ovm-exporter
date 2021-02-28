package internal

import (
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
func IOctlOCFS2Reflink(src, dst *string) error {
	params := reflinkArgs{
		OldPath:  *(*uint64)(unsafe.Pointer(src)),
		NewPath:  *(*uint64)(unsafe.Pointer(dst)),
		Preserve: 1,
	}

	fd, err := os.Open(*src)
	if err != nil {
		return errors.Wrap(err, "opening file")
	}

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd.Fd(), OCFS2IOCReflink, uintptr(unsafe.Pointer(&params))); err != 0 {
		return errors.Wrap(err, "running ioctl")
	}
	return nil
}
