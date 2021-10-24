package dgrelay

import (
	"context"
)

type Direction struct {
	RQueue *Queue
	WQueue *Queue

	A FD
	B FD

	CanRead  bool
	CanWrite bool
}

func NewDirection(afd, bfd FD, buffers, size int) *Direction {
	dir := &Direction{
		A:      afd,
		B:      bfd,
		RQueue: NewQueue(buffers),
		WQueue: NewQueue(buffers),
	}

	for i := 0; i < buffers; i += 1 {
		dir.RQueue.Add(&Buffer{
			ID:   i,
			Data: make([]byte, size),
		})
	}

	return dir
}

func (d *Direction) CanForward() bool {
	return d.CanRead && d.CanWrite || (d.CanWrite && d.WQueue.Size > 0)
}

func (d *Direction) Forward(ctx context.Context) error {
	afd := d.A
	bfd := d.B
	wqueue := d.WQueue
	rqueue := d.RQueue

	for d.CanForward() {
		if wqueue.Size > 0 {
			n, err := bfd.Write(wqueue)
			switch err {
			case nil:
				wqueue.Move(n, rqueue)

			case EWOULDBLOCK:
				d.CanWrite = false
				wqueue.Move(n, rqueue)

			default:
				return err
			}
		}

		if rqueue.Size > 0 {
			n, err := afd.Read(rqueue)
			switch err {
			case nil:
				rqueue.Move(n, wqueue)

			case EWOULDBLOCK:
				d.CanRead = false
				rqueue.Move(n, wqueue)

			default:
				return err
			}
		}
	}

	return nil
}
