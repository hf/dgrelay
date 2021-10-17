# Datagram Relay

This is a small Go library that allows transparent (and supposedly performant)
relays between a datagram-like file descriptor and a stream-like file
descriptor. It also supports relays between two datagram-like or two
stream-like file descriptors, but assuming that datagrams are being
transported. It can be used with SEQPACKET type file descriptors too.

It's useful when you want to convert a UDP socket into a TCP socket, or
transport TUN/TAP packets or frames over a TCP socket... it can be useful with
VSOCK type sockets too... or to transport QUIC over TCP.

A datagram can't be larger than 65,536 bytes.

On the stream-side datagrams are prefixed by two-bytes that indicate the size
of the datagram that follows, in big-endian order.

The relay can only forward in one direction when there is positive flow between
the sides (source can be read and sink can be written), or if there are
datagrams in the internal buffer and the sink side is ready to receive them.
This is useful because you can use the kernel's flow control algorithms on the
file descriptors (like in TCP, or UNIX sockets) to only relay when each side is
ready, allowing you to use minimal system resources and achieve maximal
throughput on the relayed link.

Note that the datagram socket *must* be both *bound* and *connected* as this
library does not care about the source or destination of datagram file
descriptors. Also, it goes without mention that the file descriptors must be
non-blocking.

## Example

```go
dgrelay.Forward(
	context.Background(),
	&dgrelay.DatagramFD{
		FD: afd,
	},
	&dgrelay.StreamFD{
		FD: bfd,
	},
	8,    /* buffers */
	2048, /* buffer size */
)
```

## License

Copyright © 2021 Stojan Dimitrovski. Some rights reserved.

Licensed under the MIT License. You can get a copy of it in `LICENSE`.
