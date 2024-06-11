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
	UserID   uint64 `json:"userID"`
	PostID   uint64 `json:"postID"`
	AuthorID uint64 `json:"authorID"`
}

func RunKafkaConsumer(topic string) {
	var consumer sarama.Consumer
	var err error
	for {
		consumer, err = sarama.NewConsumer([]string{os.Getenv("KAFKA_URL")}, sarama.NewConfig())
		if err == nil {
			break
		}
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition: %v", err)
	}
	defer partitionConsumer.Close()

	for {
		fmt.Fprintln(os.Stderr, "Receive")

		msg, ok := <-partitionConsumer.Messages()
		if !ok {
			log.Println("Channel closed")
			return
		}

		var receivedMessage kafkaMessage
		json.Unmarshal(msg.Value, &receivedMessage)

		_, err = db.Conn.Exec(
			fmt.Sprintf("INSERT INTO %v (user_id, post_id, author_id) VALUES ($1, $2, $3)", topic),
			receivedMessage.UserID,
			receivedMessage.PostID,
			receivedMessage.AuthorID,
		)
		if err != nil {
			log.Fatal("error executing a query: ", err)
		}
	}
}
