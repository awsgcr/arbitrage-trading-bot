package general

import (
	"errors"
	"jasonzhu.com/coin_labor/core/components/alerting"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// GWebsocketTimeout is an interval for sending ping/pong messages if WebsocketKeepalive is enabled
	GWebsocketTimeout = time.Second * 10
)

// WsHandler handle raw websocket message
type WsHandler func(message []byte)

type ErrHandler func(err error)

type WsServe struct {
	endpoint  string
	handler   WsHandler
	connected bool
	doneC     chan struct{}
	conn      *websocket.Conn
	closeRWM  sync.RWMutex
}

func NewWsServe(endpoint string, handler WsHandler) (*WsServe, error) {
	s := &WsServe{
		endpoint:  endpoint,
		handler:   handler,
		connected: false,
		doneC:     make(chan struct{}),
	}
	err := s.start()
	return s, err
}

func (s *WsServe) start() error {
	err := s.initConn()
	if err != nil {
		return err
	}
	s.keepalive()
	s.runReader()
	return nil
}

func (s *WsServe) initConn() error {
	Dialer := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  45 * time.Second,
		EnableCompression: false,
	}

	var err error
	s.conn, _, err = Dialer.Dial(s.endpoint, nil)
	if err != nil {
		return err
	}
	s.conn.SetReadLimit(655350)
	s.connected = true
	return nil
}

func (s *WsServe) IsClosed() bool {
	return !s.connected
}

func (s *WsServe) Close(err error) {
	s.closeRWM.Lock()
	defer func() {
		s.closeRWM.Unlock()
		alerting.NotifyRightNow(err, "websocket closed")
	}()

	if s.IsClosed() {
		return
	}

	s.connected = false
	close(s.doneC)
	s.conn.Close()
}

func (s *WsServe) DoneC() chan struct{} {
	return s.doneC
}

func (s *WsServe) runReader() {
	go func() {
		for {
			if s.IsClosed() {
				return
			}
			_, message, err := s.conn.ReadMessage()
			if err != nil {
				//fmt.Println("websocket error when read message: line 72", err)
				glg.Error("websocket error when read message: line 72", "err", err)
				s.Close(err)
				return
			}
			s.handler(message)
		}
	}()
}

func (s *WsServe) Write(m any) {
	if s.IsClosed() {
		return
	}

	err := s.conn.WriteJSON(m)
	if err != nil {
		//fmt.Println("websocket error when write message: line 85", err)
		glg.Error("websocket error when write message: line 85", "err", err)
		s.Close(err)
		return
	}
}

func (s *WsServe) keepalive() {
	var timeout = GWebsocketTimeout
	ticker := time.NewTicker(timeout)

	lastResponse := time.Now()
	s.conn.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		defer ticker.Stop()
		pingErrCnt := 0
		for {
			if s.IsClosed() {
				return
			}
			deadline := time.Now().Add(5 * time.Second)
			err := s.conn.WriteControl(websocket.PingMessage, []byte{}, deadline)
			if err != nil {
				glg.Error("websocket write pingMessage error", "err", err)
				s.Close(err)
				return
			}
			<-ticker.C
			if time.Since(lastResponse) > timeout {
				pingErrCnt++
				glg.Error("websocket ping/pong timeout", "cnt", pingErrCnt)
				if pingErrCnt >= 2 {
					s.Close(errors.New("websocket ping/pong timeout"))
					return
				}
			} else {
				pingErrCnt = 0
			}
		}
	}()
}
