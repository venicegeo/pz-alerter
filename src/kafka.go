package main

import (
	"github.com/Shopify/sarama"
)

//===========================================================================

const OffsetNewest int64 = sarama.OffsetNewest

//===========================================================================

type Kafka struct {
	host string
}

func (k *Kafka) NewProducer() (*Producer, error) {

	w := new(Producer)

	config := sarama.NewConfig()
	//config.Producer.Return.Successes = true

	producer, err := sarama.NewAsyncProducer([]string{k.host}, config)
	if err != nil {
		return nil, err
	}

	w.producer = producer
	return w, nil
}

func (k *Kafka) NewConsumer() (*Consumer, error) {
	r := new(Consumer)

	consumer, err := sarama.NewConsumer([]string{k.host}, nil)
	if err != nil {
		return nil, err
	}

	r.consumer = consumer

	return r, nil
}

//===========================================================================

type Producer struct {
	producer sarama.AsyncProducer
}

func (w *Producer) Close() error {
	return w.producer.Close()
}

// TODO: don't want to expose sarama message type, either on send or recv
func NewMessage(topic string, data string) *sarama.ProducerMessage {
	m := &sarama.ProducerMessage{Topic: topic, Key: nil, Value: sarama.StringEncoder(data)}
	return m
}

func (w *Producer) Input() chan<- *sarama.ProducerMessage {
	return w.producer.Input()
}

func (w *Producer) Successes() <-chan *sarama.ProducerMessage {
	return w.producer.Successes()
}

func (w *Producer) Errors() <-chan *sarama.ProducerError {
	return w.producer.Errors()
}

//===========================================================================

type Consumer struct {
	consumer sarama.Consumer
}

func (r *Consumer) ConsumePartition(topic string, partition int32, offset int64) (*PartitionConsumer, error) {
	spc, err := r.consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return nil, err
	}
	pc := new(PartitionConsumer)
	pc.partitionConsumer = spc
	return pc, nil
}

func (r *Consumer) Close() error {
	return r.consumer.Close()
}

//===========================================================================

type PartitionConsumer struct {
	partitionConsumer sarama.PartitionConsumer
}

func (pc *PartitionConsumer) Close() error {
	return pc.partitionConsumer.Close()
}

func (pc *PartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return pc.partitionConsumer.Messages()
}
