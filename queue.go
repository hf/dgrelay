package dgrelay

type Queue struct {
	Buffers []*Buffer
	Head    int
	Tail    int
	Size    int
}

func NewQueue(capacity int) *Queue {
	return &Queue{
		Buffers: make([]*Buffer, capacity),
	}
}

func (q *Queue) Add(buf *Buffer) {
	if q.Size == len(q.Buffers) {
		panic("queue is full")
	}

	q.Buffers[q.Tail] = buf
	q.Tail = (q.Tail + 1) % len(q.Buffers)
	q.Size += 1
}

func (q *Queue) Peek(off int) *Buffer {
	return q.Buffers[(q.Head+off)%len(q.Buffers)]
}

func (q *Queue) Pop(n int) {
	if n > q.Size {
		panic("queue has too little elements")
	}

	for i := 0; i < n; i += 1 {
		q.Buffers[(q.Head+i)%len(q.Buffers)] = nil
	}

	q.Head = (q.Head + n) % len(q.Buffers)
	q.Size -= n
}

func (q *Queue) Move(n int, to *Queue) {
	if n > q.Size {
		panic("queue has too little elements")
	}

	for i := 0; i < n; i += 1 {
		to.Add(q.Peek(i))
	}

	q.Pop(n)
}
