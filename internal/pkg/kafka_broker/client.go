package kafka_broker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

type Client struct {
	producer  sarama.SyncProducer
	reader    sarama.PartitionConsumer
	topic     string
	mu        sync.Mutex
	respChans map[string]chan *sarama.ConsumerMessage
	buffer    int
}

type Config struct {
	Brokers    []string
	Topic      string
	GroupID    string
	Partition  int32
	BufferSize int
	ReadTO     time.Duration
	WriteTO    time.Duration
}

// NewClient creates a producer and (optional) consumer on the same topic.
func NewClient(cfg Config) (*Client, error) {
	scfg := sarama.NewConfig()
	scfg.Producer.RequiredAcks = sarama.WaitForAll
	scfg.Producer.Return.Successes = true
	scfg.Net.DialTimeout = cfg.WriteTO
	scfg.Net.ReadTimeout = cfg.ReadTO

	prod, err := sarama.NewSyncProducer(cfg.Brokers, scfg)
	if err != nil {
		return nil, err
	}

	// If you only need to PRODUCE, you can skip creating the consumer.
	cons, err := sarama.NewConsumer(cfg.Brokers, scfg)
	if err != nil {
		prod.Close()
		return nil, err
	}
	partCons, err := cons.ConsumePartition(cfg.Topic, cfg.Partition, sarama.OffsetNewest)
	if err != nil {
		prod.Close()
		cons.Close()
		return nil, err
	}

	client := &Client{
		producer:  prod,
		reader:    partCons,
		topic:     cfg.Topic,
		respChans: make(map[string]chan *sarama.ConsumerMessage),
		buffer:    cfg.BufferSize,
	}

	// start dispatcher if you still do request/response:
	go client.dispatch()

	return client, nil
}

// Publish just sends a message to the single topic.
func (c *Client) Publish(_ context.Context, key string, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic:     c.topic,
		Key:       sarama.StringEncoder(key),
		Value:     sarama.ByteEncoder(b),
		Timestamp: time.Now(),
	}
	_, _, err = c.producer.SendMessage(msg)

	return err
}

// dispatch routes incoming messages (only if you need to consume in the same client).
func (c *Client) dispatch() {
	for msg := range c.reader.Messages() {
		key := string(msg.Key)
		c.mu.Lock()
		if ch, ok := c.respChans[key]; ok {
			ch <- msg
			close(ch)
			delete(c.respChans, key)
		}
		c.mu.Unlock()
	}
}

// Close shuts everything down.
func (c *Client) Close() error {
	c.producer.Close()

	return c.reader.Close()
}
