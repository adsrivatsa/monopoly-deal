package main

import (
	"context"
	stderrors "errors"
	"fmt"
	"fun-kames/internal/errors"
	"fun-kames/internal/schema"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type socket struct {
	conn    *websocket.Conn
	writeMu sync.Mutex
	writeCh chan *schema.ServerMessage
	ctx     context.Context
	cancel  context.CancelFunc
	closed  sync.Once
}

func newSocket(conn *websocket.Conn, parentCtx context.Context) (*socket, context.Context) {
	ctx, cancel := context.WithCancel(parentCtx)

	s := &socket{
		conn:    conn,
		writeCh: make(chan *schema.ServerMessage, 32),
		ctx:     ctx,
		cancel:  cancel,
	}

	go s.writeLoop()

	return s, ctx
}

func (s *socket) writeLoop() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.writeCh:
			_ = s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			data, err := proto.Marshal(msg)
			if err != nil {
				fmt.Println(err)
				continue
			}

			s.writeMu.Lock()
			err = s.conn.WriteMessage(websocket.BinaryMessage, data)
			s.writeMu.Unlock()

			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func (s *socket) close(err error) {
	s.closed.Do(func() {
		s.cancel()

		s.writeMu.Lock()
		defer s.writeMu.Unlock()

		if err != nil {
			closeMsg := websocket.FormatCloseMessage(
				websocket.CloseNormalClosure,
				err.Error(),
			)
			_ = s.conn.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(500*time.Millisecond))
		}
		_ = s.conn.Close()
	})
}

func (s *socket) send(msg *schema.ServerMessage) {
	select {
	case <-s.ctx.Done():
		return

	case s.writeCh <- msg:
	}
}

func (s *socket) read() *schema.ClientMessage {
	_, data, err := s.conn.ReadMessage()
	if err != nil {
		s.error(err)
		return nil
	}

	msg := &schema.ClientMessage{}
	err = proto.Unmarshal(data, msg)
	if err != nil {
		s.error(err)
		return nil
	}

	return msg
}

func (s *socket) error(err error) {
	var intErr errors.Error
	if !stderrors.As(err, &intErr) {
		intErr = errors.Internal(err)
	}
	s.close(intErr)
}
