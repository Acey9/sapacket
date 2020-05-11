package outputer

import (
	"github.com/Acey9/apacket/logp"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func NewKafka(broker, topic string) Outputer {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": broker,
		//"broker.version.fallback": version,
		//"api.version.request":     false,
		"retries": 8,
	}
	p, err := kafka.NewProducer(conf)
	if err != nil {
		logp.Err("%v", err)
		return nil
	}

	publisher := &Kafka{
		msgQueue:     make(chan string, 1024),
		deliveryChan: make(chan kafka.Event),
		producer:     p,
		broker:       broker,
		topic:        topic,
	}

	go publisher.Start()
	return publisher
}

type Kafka struct {
	producer     *kafka.Producer
	broker       string
	topic        string
	msgQueue     chan string
	deliveryChan chan kafka.Event
}

func (k *Kafka) Output(msg string) {
	logp.Debug("kafka", "broker:%v, topic:%v", k.broker, k.topic)
	k.msgQueue <- msg
}

func (k *Kafka) produce(msg string) {
	logging(msg)
	kmsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
		Value:          []byte(msg),
		Headers:        []kafka.Header{{Key: "UA", Value: []byte("sapacket")}},
	}
	err := k.producer.Produce(kmsg, k.deliveryChan)
	if err != nil {
		logp.Err("%v", err)
		return
	}
}

func (k *Kafka) callBack(e kafka.Event) {
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		logp.Err("delivery failed: %v", m.TopicPartition.Error)
	} else {
		logp.Debug("outputer", "delivered message to topic %s [%d] at offset %v",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}
}

func (k *Kafka) Start() {
	defer close(k.msgQueue)
	defer close(k.deliveryChan)
	defer k.producer.Close()
	for {
		select {
		case msg := <-k.msgQueue:
			k.produce(msg)
		case e := <-k.deliveryChan:
			k.callBack(e)
		}
	}
}
