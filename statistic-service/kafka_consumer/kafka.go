package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"statistic-service/db"

	"github.com/IBM/sarama"
)

type kafkaMessage struct {
	UserID uint64 `json:"userID"`
	PostID uint64 `json:"postID"`
}

func RunKafkaConsumer(topic string) {
	consumer, err := sarama.NewConsumer([]string{os.Getenv("KAFKA_URL")}, sarama.NewConfig())
	if err != nil {
		log.Fatalf("Failed to create kafka consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition: %v", err)
	}
	defer partitionConsumer.Close()

	for {
		msg, ok := <-partitionConsumer.Messages()
		if !ok {
			log.Println("Channel closed")
			return
		}

		var receivedMessage kafkaMessage
		json.Unmarshal(msg.Value, &receivedMessage)

		_, err = db.Conn.Exec(
			fmt.Sprintf("INSERT INTO %v (user_id, post_id) VALUES ($1, $2)", topic),
			receivedMessage.UserID,
			receivedMessage.PostID,
		)
		if err != nil {
			log.Fatal("error executing a query: ", err)
		}

		fmt.Fprintln(os.Stderr, "Receive")
	}
}
