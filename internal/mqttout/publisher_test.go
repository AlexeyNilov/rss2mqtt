package mqttout

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AlexeyNilov/rss2mqtt/internal/discovery"
)

func TestPublisherPublishesFormattedItem(t *testing.T) {
	client := &fakeClient{}
	publisher := NewPublisherWithClient(Config{
		BrokerURL: "tcp://localhost:1883",
		Topic:     "rss/approved",
		ClientID:  "rss2mqtt-test",
		Timeout:   time.Second,
	}, client)

	err := publisher.Publish(context.Background(), discovery.Item{
		SourceName: "sample",
		Title:      "Fighting Tool Sprawl",
		Link:       "https://example.com/article",
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if !client.connected {
		t.Fatal("client was not connected")
	}
	if !client.disconnected {
		t.Fatal("client was not disconnected")
	}
	if client.topic != "rss/approved" {
		t.Fatalf("topic = %q", client.topic)
	}
	if client.qos != 1 {
		t.Fatalf("qos = %d, want 1", client.qos)
	}
	if !strings.Contains(client.payload, "Title: Fighting Tool Sprawl") {
		t.Fatalf("payload = %q, want formatted item", client.payload)
	}
}

func TestPublisherReturnsConnectError(t *testing.T) {
	client := &fakeClient{connectErr: errors.New("connect failed")}
	publisher := NewPublisherWithClient(Config{Topic: "rss/approved", Timeout: time.Second}, client)

	err := publisher.Publish(context.Background(), discovery.Item{Title: "item"})
	if err == nil {
		t.Fatal("Publish() error = nil, want connect error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "connect") {
		t.Fatalf("Publish() error = %q, want connect context", err)
	}
}

func TestPublisherReturnsPublishError(t *testing.T) {
	client := &fakeClient{publishErr: errors.New("publish failed")}
	publisher := NewPublisherWithClient(Config{Topic: "rss/approved", Timeout: time.Second}, client)

	err := publisher.Publish(context.Background(), discovery.Item{Title: "item"})
	if err == nil {
		t.Fatal("Publish() error = nil, want publish error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "publish") {
		t.Fatalf("Publish() error = %q, want publish context", err)
	}
}

func TestPublisherReturnsTimeout(t *testing.T) {
	client := &fakeClient{connectTimeout: true}
	publisher := NewPublisherWithClient(Config{Topic: "rss/approved", Timeout: time.Second}, client)

	err := publisher.Publish(context.Background(), discovery.Item{Title: "item"})
	if err == nil {
		t.Fatal("Publish() error = nil, want timeout")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "timeout") {
		t.Fatalf("Publish() error = %q, want timeout context", err)
	}
}

type fakeClient struct {
	connected      bool
	disconnected   bool
	connectErr     error
	publishErr     error
	connectTimeout bool
	publishTimeout bool
	topic          string
	qos            byte
	payload        string
}

func (c *fakeClient) Connect() token {
	c.connected = true
	return fakeToken{wait: !c.connectTimeout, err: c.connectErr}
}

func (c *fakeClient) Publish(topic string, qos byte, _ bool, payload any) token {
	c.topic = topic
	c.qos = qos
	c.payload = string(payload.([]byte))
	return fakeToken{wait: !c.publishTimeout, err: c.publishErr}
}

func (c *fakeClient) Disconnect(uint) {
	c.disconnected = true
}

type fakeToken struct {
	wait bool
	err  error
}

func (t fakeToken) WaitTimeout(time.Duration) bool {
	return t.wait
}

func (t fakeToken) Error() error {
	return t.err
}
