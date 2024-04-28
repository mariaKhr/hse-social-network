package handlers

import (
	"encoding/json"
	kafka "main-service/kafka_producer"
	"main-service/schemas"
	"net/http"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

func ViewPost(c *gin.Context) {
	sendMessageToTopic(c, "views")
}

func LikePost(c *gin.Context) {
	sendMessageToTopic(c, "likes")
}

func sendMessageToTopic(c *gin.Context, topic string) {
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(err.Error())
		return
	}

	message := &schemas.KafkaMessage{
		PostID: uint64(postId),
		UserID: getUserID(c),
	}
	bytes, err := json.Marshal(message)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(err.Error())
		return
	}

	_, _, err = kafka.KafkaProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(bytes),
	})
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Writer.WriteString(err.Error())
		return
	}
}
