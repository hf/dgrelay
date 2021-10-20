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

		read := buf.ROff

		shouldRun := true

		for shouldRun {
			n := 0
			var err error

			if read < 2 {
				n, err = unix.Read(fd.FD, buf.Data[read:2+fd.MinData])
			} else {
				n, err = unix.Read(fd.FD, buf.Data[read:2+buf.Readsize()])
			}

			if n > -1 {
				read += n
				buf.ROff = read
			}

			switch err {
			case unix.EINTR:
				fallthrough

			case nil:
				shouldRun = read < 3 || read < (2+buf.Readsize())

				if !shouldRun {
					buf.WOff = 0 // ready for writing
				}

			default:
				return i, err
			}
		}
	}

	return qsize, nil
}

func (fd *StreamFD) Write(queue *Queue) (int, error) {
	qsize := queue.Size

	for i := 0; i < qsize; i += 1 {
		buf := queue.Peek(i)

		write := buf.WOff

		shouldRun := true

		for shouldRun {
			n, err := unix.Write(fd.FD, buf.Data[write:2+buf.Readsize()])

			if n > -1 {
				write += n
				buf.WOff = write
			}

			switch err {
			case unix.EINTR:
				shouldRun = true

			case nil:
				buf.ROff = 0 // ready for reading
				shouldRun = false

			default:
				return i, err
			}
		}
	}

	return qsize, nil
}

func (fd *StreamFD) Close() error {
	return closeFD(fd.FD)
}
