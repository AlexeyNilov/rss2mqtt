package mqttout

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/AlexeyNilov/rss2mqtt/internal/feed"
	"github.com/AlexeyNilov/rss2mqtt/internal/output"
	paho "github.com/eclipse/paho.mqtt.golang"
)

const publishQoS byte = 1

type Publisher struct {
	cfg    Config
	client mqttClient
}

type mqttClient interface {
	Connect() token
	Publish(topic string, qos byte, retained bool, payload any) token
	Disconnect(quiesce uint)
}

type token interface {
	WaitTimeout(time.Duration) bool
	Error() error
}

func NewPublisher(cfg Config) *Publisher {
	return NewPublisherWithClient(cfg, pahoClient{client: newPahoClient(cfg)})
}

func NewPublisherWithClient(cfg Config, client mqttClient) *Publisher {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}

	return &Publisher{cfg: cfg, client: client}
}

func (p *Publisher) Publish(ctx context.Context, item feed.Item) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := waitToken("connect mqtt", p.client.Connect(), p.cfg.Timeout); err != nil {
		return err
	}
	defer p.client.Disconnect(250)

	payload, err := formatPayload(item)
	if err != nil {
		return err
	}
	if err := waitToken("publish mqtt", p.client.Publish(p.cfg.Topic, publishQoS, false, payload), p.cfg.Timeout); err != nil {
		return err
	}

	return ctx.Err()
}

func newPahoClient(cfg Config) paho.Client {
	opts := paho.NewClientOptions()
	opts.AddBroker(cfg.BrokerURL)
	opts.SetClientID(cfg.ClientID)
	opts.SetAutoReconnect(false)
	opts.SetConnectRetry(false)
	opts.SetWriteTimeout(cfg.Timeout)
	return paho.NewClient(opts)
}

type pahoClient struct {
	client paho.Client
}

func (c pahoClient) Connect() token {
	return c.client.Connect()
}

func (c pahoClient) Publish(topic string, qos byte, retained bool, payload any) token {
	return c.client.Publish(topic, qos, retained, payload)
}

func (c pahoClient) Disconnect(quiesce uint) {
	c.client.Disconnect(quiesce)
}

func waitToken(action string, token token, timeout time.Duration) error {
	if !token.WaitTimeout(timeout) {
		return fmt.Errorf("%s timeout", action)
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}

	return nil
}

func formatPayload(item feed.Item) ([]byte, error) {
	var buf bytes.Buffer
	if err := output.WriteItem(&buf, item); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
