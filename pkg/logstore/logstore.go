package logstore

type Writer interface {
	Write (log string) error
}

type LogStore interface {
	Stream (writer Writer) error 
	Stop () error
	Push (log string) error
}