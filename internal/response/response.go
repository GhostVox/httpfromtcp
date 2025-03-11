package response

import (
	"fmt"
	"io"

	"github.com/GhostVox/httptcp/internal/headers"
)

type Writer struct {
	Writer      io.Writer
	WriterState WriterState
}

type WriterState struct {
	statusLineWritten bool
	headersWritten    bool
	bodyWritten       bool
}

func NewResponse(w io.Writer) Writer {
	return Writer{
		Writer: w,
		WriterState: WriterState{
			statusLineWritten: false,
			headersWritten:    false,
			bodyWritten:       false,
		},
	}
}

type StatusCode int

const (
	Success             StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case Success:
		_, err := fmt.Fprintf(w, "HTTP/1.1 200 OK\r\n")
		if err != nil {
			return err
		}
		return nil
	case BadRequest:
		_, err := fmt.Fprintf(w, "HTTP/1.1 400 Bad Request\r\n")
		if err != nil {
			return err
		}
		return nil
	case InternalServerError:
		_, err := fmt.Fprintf(w, "HTTP/1.1 500 Internal Server Error\r\n")
		if err != nil {
			return err
		}
		return nil
	default:
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d \r\n", statusCode)
		if err != nil {
			return err
		}
		return nil
	}
}

func GetDefaultHeaders(content int) headers.Headers {

	return headers.Headers{
		"Content-Type":   "text/plain",
		"Content-Length": fmt.Sprintf("%d", content),
		"Connection":     "close",
	}
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "\r\n")
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.WriterState.statusLineWritten {
		return fmt.Errorf("Status line already written")
	}
	err := WriteStatusLine(w.Writer, statusCode)
	if err != nil {
		return err
	}
	w.WriterState.statusLineWritten = true
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if !w.WriterState.statusLineWritten {
		return fmt.Errorf("Status line not written")
	}
	if w.WriterState.headersWritten {
		return fmt.Errorf("Headers already written")
	}
	err := WriteHeaders(w.Writer, headers)
	if err != nil {
		return err
	}
	w.WriterState.headersWritten = true
	return nil
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	if !w.WriterState.statusLineWritten {
		return 0, fmt.Errorf("Status line not written")
	}
	if !w.WriterState.headersWritten {
		return 0, fmt.Errorf("Headers not written")
	}
	if w.WriterState.bodyWritten {
		return 0, fmt.Errorf("Body already written")
	}
	_, err := w.Writer.Write(body)
	if err != nil {
		return 0, err
	}
	w.WriterState.bodyWritten = true
	return len(body), nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if !w.WriterState.statusLineWritten {
		return 0, fmt.Errorf("Status line not written")
	}
	if !w.WriterState.headersWritten {
		return 0, fmt.Errorf("Headers not written")
	}

	w.Writer.Write([]byte(fmt.Sprintf("%x\r\n", len(p))))
	n, err := w.Writer.Write(p)
	if err != nil {
		return n, err
	}
	n, err = w.Writer.Write([]byte("\r\n"))
	if err != nil {
		return n, err
	}
	return n, nil
}

func (w *Writer) WriteChunkedBodyEnd() error {
	if !w.WriterState.statusLineWritten {
		return fmt.Errorf("Status line not written")
	}
	if !w.WriterState.headersWritten {
		return fmt.Errorf("Headers not written")
	}

	_, err := w.Writer.Write([]byte("0\r\n"))
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteTrailers(trailers headers.Headers) error {
	if !w.WriterState.statusLineWritten {
		return fmt.Errorf("Status line not written")
	}
	if !w.WriterState.headersWritten {
		return fmt.Errorf("Headers not written")
	}
	for key, value := range trailers {
		_, err := fmt.Fprintf(w.Writer, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w.Writer, "\r\n")
	if err != nil {
		return err
	}
	return nil
}
