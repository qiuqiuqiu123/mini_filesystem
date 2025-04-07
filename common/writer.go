package common

import "os"

type Writer struct {
	WriterAt *os.File
	Offset   int64
}

func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.WriterAt.WriteAt(p, w.Offset)
	if err != nil {
		return 0, err
	}
	return n, nil
}
