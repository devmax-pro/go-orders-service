package publisher

import (
	"context"
	"fmt"
	stan "github.com/nats-io/stan.go"
	"log"
	"time"
)

const (
	defaultConnAttempts = 5
	defaultConnTimeout  = time.Second
)

type MessageHandler interface {
	Handle(context.Context, []byte)
}

type NatsPublisher struct {
	connAttempts int
	connTimeout  time.Duration
	Conn         stan.Conn
}

func New(URL, clusterId, clientId string) (*NatsPublisher, error) {
	pub := &NatsPublisher{
		connAttempts: defaultConnAttempts,
		connTimeout:  defaultConnTimeout,
	}

	var err error
	for pub.connAttempts > 0 {
		pub.Conn, err = stan.Connect(
			clusterId,
			clientId,
			stan.NatsURL("nats://"+URL),
		)
		if err == nil {
			break
		}

		log.Println(fmt.Sprintf("NatsPublisher is trying to connect to NATS Streaming, attempts left: %d", pub.connAttempts))
		time.Sleep(pub.connTimeout)
		pub.connAttempts--
	}

	if err != nil {
		log.Println("Connection to NATS Streaming failed, attempts are over: ", err)
		return nil, err
	}

	log.Println(fmt.Sprintf("Connected to %s clusterID: [%s] clientID: [%s]\n", URL, clusterId, clientId))

	return pub, nil
}

func (s *NatsPublisher) Close() error {
	if s.Conn != nil {
		return s.Close()
	}
	return nil
}

func (s *NatsPublisher) Publish(channel string, msg []byte) error {

	err := s.Conn.Publish(channel, msg)

	if err != nil {
		log.Println("Publishing message failed on channel: " + channel)
		return err
	}
	log.Println("Published message on channel: " + channel)
	return nil
}
