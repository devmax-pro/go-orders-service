package subscriber

import (
	"context"
	"fmt"
	"time"

	logs "github.com/devmax-pro/order-service/internal/adapters/logger"
	stan "github.com/nats-io/stan.go"
)

const (
	clusterID           = "orders-streaming"
	clientID            = "order-service-client"
	durableID           = "order-service-durable"
	defaultConnAttempts = 5
	defaultConnTimeout  = time.Second
)

type MessageHandler interface {
	Handle(context.Context, []byte) error
}

type NatsSubscriber struct {
	connAttempts int
	connTimeout  time.Duration
	Conn         stan.Conn
}

func New(URL string) (*NatsSubscriber, error) {
	sb := &NatsSubscriber{
		connAttempts: defaultConnAttempts,
		connTimeout:  defaultConnTimeout,
	}

	var err error
	for sb.connAttempts > 0 {
		sb.Conn, err = stan.Connect(
			clusterID,
			clientID,
			stan.NatsURL("nats://"+URL),
		)
		if err == nil {
			break
		}

		logs.Error(fmt.Sprintf("NatsSubscriber is trying to connect to NATS Streaming, attempts left: %d", sb.connAttempts))
		time.Sleep(sb.connTimeout)
		sb.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("connection to NATS Streaming failed, attempts are over: %w", err)
	}

	logs.Info(fmt.Sprintf("Connected to %s clusterID: [%s] clientID: [%s]\n", URL, clusterID, clientID))

	return sb, nil
}

func (s *NatsSubscriber) Subscribe(channel string, handler MessageHandler) error {

	aw, _ := time.ParseDuration("60s")
	_, err := s.Conn.Subscribe(channel, func(m *stan.Msg) {
		err := m.Ack() // Manual Ack
		if err != nil {
			logs.Error("Message ack failed", err)
			return
		}
		logs.Infof("Received a message: %s from channel: %s", string(m.Data), channel)
		err = handler.Handle(context.Background(), m.Data)
		if err != nil {
			logs.Error("Message handling failed", err)
			return
		}
		logs.Info("Message handling successful")
	}, stan.DurableName(durableID),
		stan.MaxInflight(25),
		stan.SetManualAckMode(),
		stan.AckWait(aw),
	)

	if err != nil {
		return fmt.Errorf("subscription failed: %w", err)
	}

	logs.Info("Subscription successful")
	return nil
}

func (s *NatsSubscriber) Close() error {
	if s.Conn != nil {
		return s.Close()
	}
	return nil
}
