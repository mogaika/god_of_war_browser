package status

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	INFO = iota
	ERROR
	PROGRESS
)

type status struct {
	Message  string
	Time     time.Time
	Type     int
	Progress float32
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

func (c *client) writePump() {
	ticker := time.NewTicker(time.Second * 30)
	defer func() {
		unregisterClient(c)
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(40 * time.Second)); err != nil {
				panic(err)
			}
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("[status] ws write msg error: %v", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(40 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[status] ws write ping error: %v", err)
				return
			}
		}
	}
}

func NewClient(conn *websocket.Conn) *client {
	c := &client{conn: conn, send: make(chan []byte, 32)}
	registerClient(c)
	go c.writePump()
	globalLock.Lock()
	defer globalLock.Unlock()
	c.send <- lastMessage
	return c
}

var statusBroadcast chan *status
var broadcastList map[*client]bool
var globalLock sync.Mutex
var lastMessage []byte = nil

func registerClient(c *client) {
	globalLock.Lock()
	defer globalLock.Unlock()
	broadcastList[c] = true
}

func unregisterClient(c *client) {
	globalLock.Lock()
	defer globalLock.Unlock()
	delete(broadcastList, c)
}

func init() {
	statusBroadcast = make(chan *status, 16)
	broadcastList = make(map[*client]bool)
	go func() {
		for {
			select {
			case s := <-statusBroadcast:
				data, err := json.Marshal(s)
				if err != nil {
					panic(err)
				}
				globalLock.Lock()
				lastMessage = data
				for c := range broadcastList {
					c.send <- data
				}
				globalLock.Unlock()
			}
		}
	}()
}

func Status(msg string, _type int, progress float32) {
	if math.IsNaN(float64(progress)) || math.IsInf(float64(progress), 0) {
		progress = 0
	}
	statusBroadcast <- &status{
		Message:  msg,
		Time:     time.Now(),
		Type:     _type,
		Progress: progress}
}

func Info(format string, a ...interface{}) {
	Status(fmt.Sprintf(format, a...), INFO, 0.0)
}

func Error(format string, a ...interface{}) {
	Status(fmt.Sprintf(format, a...), ERROR, 0.0)
}

func Progress(progress float32, format string, a ...interface{}) {
	Status(fmt.Sprintf(format, a...), PROGRESS, progress)
}
