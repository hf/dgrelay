package dgrelay

import (
	"golang.org/x/sys/unix"
)

type StreamFD struct {
	FD      int
	MinData int
}

func (fd *StreamFD) Unix() int {
	return fd.FD
}

func (fd *StreamFD) Read(queue *Queue) (int, error) {
	qsize := queue.Size

	for i := 0; i < qsize; i += 1 {
		buf := queue.Peek(i)

		roff := buf.ROff

	read: // goto label
		n := 0
		var err error

		if roff < 2 {
			n, err = unix.Read(fd.FD, buf.Data[roff:2+fd.MinData])
		} else {
			n, err = unix.Read(fd.FD, buf.Data[roff:2+buf.Readsize()])
		}

		if n > -1 {
			roff += n
			buf.ROff = roff
		}

		switch err {
		case unix.EINTR:
			goto read

		case nil:
			if roff < 3 {
				goto read
			}

			buf.WOff = 0 // ready for writing

		default:
			return i, err
		}
	}

	return qsize, nil
}

func (fd *StreamFD) Write(queue *Queue) (int, error) {
	qsize := queue.Size

	for i := 0; i < qsize; i += 1 {
		buf := queue.Peek(i)

		woff := buf.WOff

	write: // goto label
		n, err := unix.Write(fd.FD, buf.Data[woff:2+buf.Readsize()])

		if n > -1 {
			woff += n
			buf.WOff = woff
		}

		switch err {
		case unix.EINTR:
			goto write

		case nil:
			buf.ROff = 0 // ready for reading

		default:
			return i, err
		}
	}

	return qsize, nil
}

func (fd *StreamFD) Close() error {
	return closeFD(fd.FD)
}
