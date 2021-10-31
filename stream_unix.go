package dgrelay

import (
	"fmt"
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
		mindata := 2 + fd.MinData

		roff := buf.ROff
		if roff >= mindata {
			mindata = 2 + buf.Readsize()
		}

	read: // goto label
		n, err := unixRead(fd.FD, buf.Data[roff:mindata])

		if n > -1 {
			roff += n
			buf.ROff = roff

			if roff >= mindata {
				mindata = 2 + buf.Readsize()

				if mindata > len(buf.Data) {
					return i, fmt.Errorf("packet signals %v bytes to receive but buffer can only hold %v bytes", mindata, len(buf.Data))
				}
			}
		}

		switch err {
		case unixEINTR:
			goto read

		case nil:
			if roff < mindata {
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
		bufsize := 2 + buf.Readsize()

		woff := buf.WOff

	write: // goto label
		n, err := unixWrite(fd.FD, buf.Data[woff:bufsize])

		if n > -1 {
			woff += n
			buf.WOff = woff
		}

		switch err {
		case unixEINTR:
			goto write

		case nil:
			if woff < bufsize {
				goto write
			}

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
