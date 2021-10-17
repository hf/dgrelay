package dgrelay

type Buffer struct {
	ID   int
	Data []byte
	ROff int
	WOff int
}

func (buf *Buffer) Writesize(size int) {
	buf.Data[0] = byte((size >> 8) & 0xFF)
	buf.Data[1] = byte((size >> 0) & 0xFF)
}

func (buf *Buffer) Readsize() int {
	return (int(buf.Data[0]) << 8) | int(buf.Data[1])
}
