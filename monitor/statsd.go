package monitor

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

var (
	ErrNoConnection = errors.New("no connection!")
)

type StatsD struct {
	conn      net.Conn
	c         chan string
	Addr      string
	BatchSize int64
	Name      string
}

func NewStatsD(name, addr string, batchSize int64) *StatsD {
	return &StatsD{
		Addr:      addr,
		Name:      name,
		BatchSize: batchSize,
		c:         make(chan string, batchSize),
	}
}

func (s *StatsD) Dail() error {
	addr, err := net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		return err
	}
	s.conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	go func() {
		leftSize := 0
		var contents []string
		for content := range s.c {
			contents = append(contents, content)
			if leftSize <= 0 {
				leftSize = len(s.c)
			}

			if leftSize > 0 {
				leftSize--
			} else {
				content = strings.Join(contents, "\n")
				if err := s.send(content); err != nil {
					log.Println("[statsd] send Err", err)
				}
				contents = []string{}
			}
		}
	}()
	return nil
}

func (s *StatsD) Close() error {
	close(s.c)
	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		return err
	}
	return nil
}

func (s *StatsD) Count(key string, value int64) {
	s.ch(fmt.Sprintf("%s.%s:%d|c", s.Name, key, value))
}

func (s *StatsD) Time(key string, value int64) {
	s.ch(fmt.Sprintf("%s.%s:%d|ms", s.Name, key, value))
}

func (s *StatsD) Gauge(key string, value int64) {
	s.ch(fmt.Sprintf("%s.%s:%d|g", s.Name, key, value))
}

func (s *StatsD) GaugeDiff(key string, value int64) {
	strVar := ""
	if value >= 0 {
		strVar = fmt.Sprintf("+%d", value)
	} else {
		strVar = fmt.Sprintf("%d", value)
	}
	s.ch(fmt.Sprintf("%s.%s:%s|g", s.Name, key, strVar))
}

func (s *StatsD) send(content string) error {
	if s.conn == nil {
		return ErrNoConnection
	}
	s.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 10))
	_, err := s.conn.Write([]byte(content))
	return err
}

func (s *StatsD) ch(str string) {
	select {
	case s.c <- str:
		return
	default:
		log.Println("[Warning] statsD send channel is full!")
	}
}
