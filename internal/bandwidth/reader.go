package bandwidth

import "io"

type Reader struct {
	reader io.Reader
}

func WrapReader(r io.Reader) *Reader {
	return &Reader{reader: r}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if n > 0 {
		AddIn(int64(n))
	}
	return n, err
}

type Writer struct {
	writer io.Writer
}

func WrapWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.writer.Write(p)
	if n > 0 {
		AddOut(int64(n))
	}
	return n, err
}