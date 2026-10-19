package ws

import (
	"context"
	"io"
	"sync"

	"github.com/coder/websocket"
)

// NewWrapper returns a new ReadWriteCloser wrapper for a websocket connection.
func NewWrapper(ctx context.Context, conn *websocket.Conn) io.ReadWriteCloser {
	return &wrapper{ctx: ctx, conn: conn}
}

// wrapper implements ReadWriteCloser on top of a websocket connection.
type wrapper struct {
	ctx    context.Context
	conn   *websocket.Conn
	reader io.Reader
	mur    sync.Mutex
	muw    sync.Mutex
}

func (w *wrapper) Read(p []byte) (n int, err error) {
	w.mur.Lock()
	defer w.mur.Unlock()

	if w.reader == nil {
		var mt websocket.MessageType

		mt, w.reader, err = w.conn.Reader(w.ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				return 0, io.EOF
			}

			return 0, err
		}
		if mt == websocket.MessageText {
			w.reader = nil // At the end of the message, reset the reader.
			return 0, io.EOF
		}
	}

	// Perform the read itself.
	n, err = w.reader.Read(p)
	if err != nil {
		w.reader = nil // At the end of the message, reset the reader.

		if err == io.EOF {
			return n, nil // Don't return EOF error at end of message.
		}

		return n, err
	}

	return n, nil
}

func (w *wrapper) Write(p []byte) (int, error) {
	w.muw.Lock()
	defer w.muw.Unlock()

	// Send the data as a binary message.
	err := w.conn.Write(w.ctx, websocket.MessageBinary, p)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// Close sends a control message indicating the stream is finished, but it does not close the socket.
func (w *wrapper) Close() error {
	w.muw.Lock()
	defer w.muw.Unlock()
	return w.conn.Close(websocket.StatusNormalClosure, "")
}
