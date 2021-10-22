package dgrelay

import (
	"bytes"
	"context"
	tst "testing"
	"time"

	"golang.org/x/sys/unix"
)

func createSocketpair(domain, typ int) ([2]int, error) {
socketpair: // goto label
	pair, err := unix.Socketpair(domain, typ|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)

	switch err {
	case unix.EINTR:
		goto socketpair

	case nil:
		return pair, nil

	default:
		return [2]int{-1, -1}, err
	}
}

func TestDatagramDatagram(t *tst.T) {
	apair, err := createSocketpair(unix.AF_UNIX, unix.SOCK_DGRAM)
	if nil != err {
		t.Fatal(err)
	}

	bpair, err := createSocketpair(unix.AF_UNIX, unix.SOCK_DGRAM)
	if nil != err {
		t.Fatal(err)
	}

	go Forward(context.Background(), &DatagramFD{
		FD: apair[1],
	}, &DatagramFD{
		FD: bpair[0],
	}, 1, 2048)

	exampleBytes := []byte{1, 2, 3, 4, 5, 6, 7}

writeA: // goto label
	n, err := unix.Write(apair[0], exampleBytes)
	switch err {
	case unix.EINTR:
		goto writeA

	case unix.EWOULDBLOCK:
		time.Sleep(10 * time.Millisecond)
		goto writeA

	case nil:
		// do nothing

	default:
		t.Fatalf("write failed %v", err)
	}

	bufA := make([]byte, 2048)

readB: // goto label
	n, err = unix.Read(bpair[1], bufA)
	switch err {
	case unix.EINTR:
		goto readB

	case unix.EWOULDBLOCK:
		time.Sleep(10 * time.Millisecond)
		goto readB

	case nil:
		if !bytes.Equal(exampleBytes, bufA[:n]) {
			t.Fatalf("read buffer is wrong, got %v expected %v", bufA[:n], exampleBytes)
		}

	default:
		t.Fatalf("read failed %v", err)
	}

	bufB := make([]byte, 2048)

writeB: // goto label
	n, err = unix.Write(bpair[1], exampleBytes)
	switch err {
	case unix.EINTR:
		goto writeB

	case unix.EWOULDBLOCK:
		time.Sleep(10 * time.Millisecond)
		goto writeB

	case nil:
		// do nothing

	default:
		t.Fatalf("write failed %v", err)
	}

readA: // goto label
	n, err = unix.Read(apair[0], bufB)
	switch err {
	case unix.EINTR:
		goto readA

	case unix.EWOULDBLOCK:
		time.Sleep(10 * time.Millisecond)
		goto readA

	case nil:
		if !bytes.Equal(exampleBytes, bufB[:n]) {
			t.Fatalf("read buffer is wrong, got %v expected %v", bufB[:n], exampleBytes)
		}

	default:
		t.Fatalf("read failed %v", err)
	}
}

func TestDatagramStream(t *tst.T) {
	apair, err := createSocketpair(unix.AF_UNIX, unix.SOCK_DGRAM)
	if nil != err {
		t.Fatal(err)
	}

	bpair, err := createSocketpair(unix.AF_UNIX, unix.SOCK_STREAM)
	if nil != err {
		t.Fatal(err)
	}

	go Forward(context.Background(), &DatagramFD{
		FD: apair[1],
	}, &StreamFD{
		FD: bpair[0],
	}, 1, 2048)

	exampleBytes := []byte{0, 7, 1, 2, 3, 4, 5, 6, 7}

writeA: // goto label
	n, err := unix.Write(apair[0], exampleBytes[2:])
	switch err {
	case unix.EINTR:
		goto writeA

	case unix.EWOULDBLOCK:
		time.Sleep(10 * time.Millisecond)
		goto writeA

	case nil:
		// do nothing

	default:
		t.Fatalf("write failed %v", err)
	}

	bufA := make([]byte, 2048)

readB: // goto label
	n, err = unix.Read(bpair[1], bufA)
	switch err {
	case unix.EINTR:
		goto readB

	case unix.EWOULDBLOCK:
		time.Sleep(10 * time.Millisecond)
		goto readB

	case nil:
		if !bytes.Equal(exampleBytes, bufA[:n]) {
			t.Fatalf("read buffer is wrong, got %v expected %v", bufA[:n], exampleBytes)
		}

	default:
		t.Fatalf("read failed %v", err)
	}

	bufB := make([]byte, 2048)

writeB: // goto label
	n, err = unix.Write(bpair[1], exampleBytes)
	switch err {
	case unix.EINTR:
		goto writeB

	case unix.EWOULDBLOCK:
		time.Sleep(10 * time.Millisecond)
		goto writeB

	case nil:
		// do nothing

	default:
		t.Fatalf("write failed %v", err)
	}

readA: // goto label
	n, err = unix.Read(apair[0], bufB)
	switch err {
	case unix.EINTR:
		goto readA

	case unix.EWOULDBLOCK:
		time.Sleep(10 * time.Millisecond)
		goto readA

	case nil:
		if !bytes.Equal(exampleBytes[2:], bufB[:n]) {
			t.Fatalf("read buffer is wrong, got %v expected %v", bufB[:n], exampleBytes)
		}

	default:
		t.Fatalf("read failed %v", err)
	}
}
