package dgrelay

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sys/unix"
)

func epollCreate() (int, error) {
	for {
		epfd, err := unix.EpollCreate(2)
		switch err {
		case unix.EINTR:
			continue

		default:
			return epfd, err
		}
	}
}

func epollAdd(epfd int, fd FD, event *unix.EpollEvent) error {
	for {
		err := unix.EpollCtl(epfd, unix.EPOLL_CTL_ADD, fd.Unix(), event)
		switch err {
		case unix.EINTR:
			continue

		default:
			return err
		}
	}
}

func epollWait(epfd int, events []unix.EpollEvent, msec int) (int, error) {
	for {
		n, err := unix.EpollWait(epfd, events, msec)
		switch err {
		case unix.EINTR:
			continue

		default:
			return n, err
		}
	}
}

func Forward(ctx context.Context, afd, bfd FD, buffers, bufferSize int) error {
	epfd, err := epollCreate()
	if nil != err {
		return err
	}

	defer closeFD(epfd)

	err = epollAdd(epfd, afd, &unix.EpollEvent{
		Events: unix.EPOLLET | unix.EPOLLIN | unix.EPOLLOUT,
		Fd:     int32(afd.Unix()),
	})
	if nil != err {
		return err
	}

	err = epollAdd(epfd, bfd, &unix.EpollEvent{
		Events: unix.EPOLLET | unix.EPOLLIN | unix.EPOLLOUT,
		Fd:     int32(bfd.Unix()),
	})
	if nil != err {
		return err
	}

	left := NewDirection(bfd, afd, buffers, bufferSize)
	right := NewDirection(afd, bfd, buffers, bufferSize)

	events := make([]unix.EpollEvent, 2)

	wg := &sync.WaitGroup{}

	for {
		n := 0

		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			n, err = epollWait(epfd, events, 150)
		}

		for i := 0; i < n; i += 1 {
			switch events[i].Fd {
			case int32(afd.Unix()):
				if 0 != (events[i].Events & unix.EPOLLIN) {
					// a can read, i.e. a -> b, i.e. right
					right.CanRead = true
				}

				if 0 != (events[i].Events & unix.EPOLLOUT) {
					// a can write, i.e. a <- b, i.e. left
					left.CanWrite = true
				}

			case int32(bfd.Unix()):
				if 0 != (events[i].Events & unix.EPOLLIN) {
					// b can read, i.e. a <- b, i.e. left
					left.CanRead = true
				}

				if 0 != (events[i].Events & unix.EPOLLOUT) {
					// b can write, i.e. a -> b, i.e. right
					right.CanWrite = true
				}

			default:
				panic(fmt.Errorf("unknown fd in epoll event %d", events[i].Fd))
			}
		}

		if left.CanForward() {
			wg.Add(1)
			go left.Forward(ctx, wg)
		}

		if right.CanForward() {
			wg.Add(1)
			go right.Forward(ctx, wg)
		}

		wg.Wait()
	}
}
