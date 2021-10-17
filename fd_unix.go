package dgrelay

import (
	"golang.org/x/sys/unix"
)

var (
	EWOULDBLOCK = unix.EWOULDBLOCK
)

func closeFD(fd int) error {
	for {
		err := unix.Close(fd)
		switch err {
		case unix.EINTR:
			continue

		default:
			return err
		}
	}
}
