package logstore

type Writer interface {
	Write(log string) error
}

type LogStore interface {
	Stream(writer Writer, stopCh <-chan struct{}) error
	Push(log string) error
}
