package kafka

import (
	"log"
	"os"

	"github.com/IBM/sarama"
)

var KafkaProducer sarama.SyncProducer

func InitKafkaProducer() {
	var err error
	KafkaProducer, err = sarama.NewSyncProducer([]string{os.Getenv("KAFKA_URL")}, nil)
	if err != nil {
		log.Fatalf("Failed to create kafka producer: %v", err)
	}
}

func CloseKafkaProducer() {
	KafkaProducer.Close()
}
