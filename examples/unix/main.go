package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sys/unix"

	"github.com/hf/dgrelay"
)

func createSocketPair(domain, typ, proto int) ([]int, error) {
	for {
		fd, err := unix.Socketpair(domain, typ, proto)
		switch err {
		case unix.EINTR:
			continue

		default:
			return fd[:], err
		}
	}
}

func main() {
	fds, err := createSocketPair(unix.AF_UNIX, unix.SOCK_DGRAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, 0)
	if nil != err {
		panic(err)
	}

	afd := &dgrelay.DatagramFD{
		FD: fds[1],
	}
	dfd := fds[0]

	fds, err = createSocketPair(unix.AF_UNIX, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, 0)
	if nil != err {
		panic(err)
	}

	bfd := &dgrelay.StreamFD{
		FD: fds[1],
	}
	sfd := fds[0]

	go func() {
		dgrelay.Forward(context.Background(), afd, bfd)
	}()

	go func() {
		buf := make([]byte, 2048)

		for {
			shouldRun := true
			for shouldRun {
				n, err := unix.Read(dfd, buf)
				switch err {
				case unix.EINTR:
					shouldRun = true

				case nil:
					fmt.Printf("afd <- %q\n", buf[:n])
					shouldRun = false

				case unix.EWOULDBLOCK:
					time.Sleep(200 * time.Millisecond)
					shouldRun = true

				default:
					panic(err)
				}
			}
		}
	}()

	go func() {
		buf := make([]byte, 2048)

		read := 0

		for {
			shouldRun := true
			for shouldRun {
				n := 0
				var err error

				if read < 2 {
					n, err = unix.Read(sfd, buf[read:2])
				} else {
					size := (int(buf[0]) << 8) | int(buf[1])
					n, err = unix.Read(sfd, buf[read:2+size])
				}

				if n > -1 {
					read += n
				}

				switch err {
				case unix.EINTR:
					shouldRun = true

				case nil:
					if read > 2 {
						size := (int(buf[0]) << 8) | int(buf[1])

						if read == size+2 {
							fmt.Printf("bfd <- %d %q\n", read, buf[2:size])
							read = 0
						}
					}

					shouldRun = false

				case unix.EWOULDBLOCK:
					time.Sleep(200 * time.Millisecond)
					shouldRun = true

				default:
					panic(err)
				}
			}
		}
	}()

	go func() {
		buf := make([]byte, 16)

		for {
			shouldRun := true

			for shouldRun {
				n, err := unix.Write(dfd, buf)
				switch err {
				case unix.EINTR:
					shouldRun = true

				case nil:
					if n != len(buf) {
						panic("here")
					}
					shouldRun = false

				case unix.EWOULDBLOCK:
					time.Sleep(200 * time.Millisecond)
					shouldRun = true

				default:
					panic(err)
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		buf := make([]byte, 16)
		buf[1] = byte(len(buf) - 2)

		written := 0

		for {
			shouldRun := true

			for shouldRun {
				n, err := unix.Write(sfd, buf[written:])

				if n > -1 {
					written += n
				}

				switch err {
				case unix.EINTR:
					shouldRun = true

				case nil:
					if written == len(buf) {
						written = 0
					}

					shouldRun = false

				case unix.EWOULDBLOCK:
					time.Sleep(200 * time.Millisecond)
					shouldRun = true

				default:
					panic(err)
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	wg.Wait()
}
