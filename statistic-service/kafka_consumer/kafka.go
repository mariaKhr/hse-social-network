package kafka

import (
	"encoding/json"
	"log"
	"os"
	"statistic-service/db"

	"github.com/IBM/sarama"
)

type kafkaMessage struct {
	UserID uint64 `json:"userID"`
	PostID uint64 `json:"postID"`
}

func RunKafkaConsumer() {
	consumer, err := sarama.NewConsumer([]string{os.Getenv("KAFKA_URL")}, nil)
	if err != nil {
		log.Fatalf("Failed to create kafka consumer: %v", err)
	}
	defer consumer.Close()

	likeConsumer, err := consumer.ConsumePartition("like", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition: %v", err)
	}
	defer likeConsumer.Close()

	viewConsumer, err := consumer.ConsumePartition("view", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition: %v", err)
	}
	defer viewConsumer.Close()

	for {
		select {
		case likeMsg, ok := <-likeConsumer.Messages():
			if !ok {
				log.Println("Likes channel closed")
				return
			}

			var receivedMessage kafkaMessage
			json.Unmarshal(likeMsg.Value, &receivedMessage)

			db.Conn.Exec(
				"INSERT INTO likes (user_id, post_id) VALUES ($1, $2)",
				receivedMessage.UserID,
				receivedMessage.PostID,
			)

		case viewMsg, ok := <-viewConsumer.Messages():
			if !ok {
				log.Println("Views channel closed")
				return
			}

			var receivedMessage kafkaMessage
			json.Unmarshal(viewMsg.Value, &receivedMessage)

			db.Conn.Exec(
				"INSERT INTO views (user_id, post_id) VALUES ($1, $2)",
				receivedMessage.UserID,
				receivedMessage.PostID,
			)
		}
	}
}
