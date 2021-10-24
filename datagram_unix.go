package dgrelay

type DatagramFD struct {
	FD int
}

func (fd *DatagramFD) Unix() int {
	return fd.FD
}

func (fd *DatagramFD) Read(queue *Queue) (int, error) {
	qsize := queue.Size

	for i := 0; i < qsize; i += 1 {
		buf := queue.Peek(i)

		if 0 != buf.ROff {
			panic("roffset not 0")
		}

	read: // goto label
		n, err := unixRead(fd.FD, buf.Data[2:])
		switch err {
		case unixEINTR:
			goto read

		case nil:
			buf.Writesize(n)

			buf.WOff = 0 // ready for writing

		default:
			return i, err
		}
	}

	return qsize, nil
}

func (fd *DatagramFD) Write(queue *Queue) (int, error) {
	qsize := queue.Size

	for i := 0; i < qsize; i += 1 {
		buf := queue.Peek(i)

		if 0 != buf.WOff {
			panic("woffset not 0")
		}

	write: // goto label
		n, err := unixWrite(fd.FD, buf.Data[2:2+buf.Readsize()])
		switch err {
		case unixEINTR:
			goto write

		case nil:
			if n != buf.Readsize() {
				panic("written bytes != buffer bytes")
			}

			buf.ROff = 0 // ready for reading

		default:
			return i, err
		}
	}

	return qsize, nil
}

func (fd *DatagramFD) Close() error {
	return closeFD(fd.FD)
}
