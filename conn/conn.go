package conn

import "bytes"

type Command func(*bytes.Buffer) error

type Conn interface {
	WriteCommand(Command) (int64, error)
}
