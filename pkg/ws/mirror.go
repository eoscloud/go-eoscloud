package ws

import (
	"context"
	"io"

	"github.com/coder/websocket"
)

func Mirror(ctx context.Context, conn *websocket.Conn, rwc io.ReadWriteCloser) (chan error, chan error) {
	chRead := mirrorRead(ctx, conn, rwc)
	chWrite := mirrorWrite(ctx, conn, rwc)
	return chRead, chWrite
}

func mirrorRead(ctx context.Context, conn *websocket.Conn, rc io.Reader) chan error {
	chDone := make(chan error, 1)
	if rc == nil {
		close(chDone)
		return chDone
	}

	connRWC := NewWrapper(ctx, conn)

	go func() {
		_, err := io.Copy(connRWC, rc)
		_ = connRWC.Close()
		chDone <- err
		close(chDone)
	}()

	return chDone
}

func mirrorWrite(ctx context.Context, conn *websocket.Conn, wc io.Writer) chan error {
	chDone := make(chan error, 1)
	if wc == nil {
		close(chDone)
		return chDone
	}

	connRWC := NewWrapper(ctx, conn)

	go func() {
		_, err := io.Copy(wc, connRWC)
		chDone <- err
		close(chDone)
	}()

	return chDone
}
