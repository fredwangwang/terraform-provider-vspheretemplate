package archive

import (
	"io"
)

type Archive interface {
	Open(string) (io.ReadCloser, int64, error)
}
