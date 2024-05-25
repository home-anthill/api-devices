package testutils

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

// This is out mqtt.Client struct mock!
// It must implement all methods of mqtt.Client to
// implicitly implement that interface
type mqttClientMock struct{}

// NewMockClient Exposes a builder function to instantiate the
// mock and return it as mqtt.Client interface
func NewMockClient() mqtt.Client {
	c := &mqttClientMock{}
	return c
}

// **************************************************
// implement all methods of mqttClientMock to match the mqtt.Client interface
// **************************************************

func (c *mqttClientMock) IsConnected() bool {
	return true
}
func (c *mqttClientMock) IsConnectionOpen() bool {
	return true
}
func (c *mqttClientMock) Connect() mqtt.Token {
	t := newToken()
	go func() {
		t.release()
	}()
	return t
}
func (c *mqttClientMock) Disconnect(quiesce uint)                             {}
func (c *mqttClientMock) AddRoute(topic string, callback mqtt.MessageHandler) {}
func (c *mqttClientMock) OptionsReader() mqtt.ClientOptionsReader {
	return mqtt.ClientOptionsReader{}
}
func (c *mqttClientMock) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	t := newToken()
	go func() {
		t.release()
	}()
	return t
}
func (c *mqttClientMock) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	t := newToken()
	return t
}
func (c *mqttClientMock) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	t := newToken()
	return t
}
func (c *mqttClientMock) Unsubscribe(topics ...string) mqtt.Token {
	t := newToken()
	return t
}

// **************************************************
// Create a Token mock with all required methods
// **************************************************

type token struct {
	err  error
	done chan struct{}
}

func newToken() *token {
	return &token{
		done: make(chan struct{}),
	}
}

func (t *token) release() {
	close(t.done)
}

func (t *token) Wait() bool {
	<-t.done
	return true
}

func (t *token) WaitTimeout(d time.Duration) bool {
	select {
	case <-t.done:
		return true
	case <-time.After(d):
		return false
	}
}

func (t *token) Done() <-chan struct{} {
	return t.done
}

func (t *token) Error() error {
	return t.err
}
