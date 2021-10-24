package dgrelay

import (
	"golang.org/x/sys/unix"
)

var (
	EWOULDBLOCK     = unix.EWOULDBLOCK
	unixEINTR       = unix.EINTR
	unixWrite       = unix.Write
	unixRead        = unix.Read
	unixClose       = unix.Close
	unixEpollCreate = unix.EpollCreate
	unixEpollCtl    = unix.EpollCtl
	unixEpollWait   = unix.EpollWait
)

func closeFD(fd int) error {
	for {
		err := unixClose(fd)
		switch err {
		case unixEINTR:
			continue

		default:
			return err
		}
	}
}
