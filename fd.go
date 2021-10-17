package dgrelay

type FD interface {
	Unix() int

	Read(*Queue) (int, error)
	Write(*Queue) (int, error)

	Close() error
}
