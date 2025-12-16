package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type SafeConn struct {
	*websocket.Conn
	mu sync.Mutex
}

func (c *SafeConn) WriteJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteJSON(v)
}

func (c *SafeConn) WriteMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteMessage(messageType, data)
}

func (c *SafeConn) WriteControl(messageType int, data []byte, deadline time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteControl(messageType, data, deadline)
}
