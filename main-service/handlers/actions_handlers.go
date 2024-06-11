package handlers

import (
	"context"
	"encoding/json"
	grpcclient "main-service/grpc_client"
	kafka "main-service/kafka_producer"
	pb "main-service/protos"
	"main-service/schemas"
	"net/http"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
)

func ViewPost(c *gin.Context) {
	sendMessageToTopic(c, "views")
}

func LikePost(c *gin.Context) {
	sendMessageToTopic(c, "likes")
}

func sendMessageToTopic(c *gin.Context, topic string) {
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(err.Error())
		return
	}

	authorID, err := getAuthorID(uint64(postID))
	if err != nil {
		c.Status(http.StatusNotFound)
		c.Writer.WriteString(err.Error())
		return
	}

	message := &schemas.KafkaMessage{
		UserID:   getUserID(c),
		PostID:   uint64(postID),
		AuthorID: authorID,
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

func GetStatisticsByPost(c *gin.Context) {
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(err.Error())
		return
	}

	getStatReq := pb.GetStatisticsRequest{
		PostID: uint64(postId),
	}

	resp, err := grpcclient.StatService.GetStatistics(context.Background(), &getStatReq)

	type Stat struct {
		PostID   uint64 `json:"postID"`
		NumLikes uint64 `json:"likes"`
		NumViews uint64 `json:"views"`
	}

	stat := Stat{
		PostID:   resp.PostID,
		NumLikes: resp.NumLikes,
		NumViews: resp.NumViews,
	}

	processServerResponse(c, err, stat)
}

func GetTop5Posts(c *gin.Context) {
	orderBy, ok := c.GetQuery("orderBy")
	if !ok {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString("no orderBy in query")
		return
	}

	var comp pb.Comparator
	switch orderBy {
	case "likes":
		comp = pb.Comparator_LIKES
	case "views":
		comp = pb.Comparator_VIEWS
	}

	getTopPostsReq := pb.GetTop5PostsRequest{
		Comparator: comp,
	}

	resp, err := grpcclient.StatService.GetTop5Posts(context.Background(), &getTopPostsReq)

	type StatLikes struct {
		PostID   uint64 `json:"postID"`
		Login    string `json:"login"`
		NumLikes uint64 `json:"likes"`
	}

	type StatViews struct {
		PostID   uint64 `json:"postID"`
		Login    string `json:"login"`
		NumViews uint64 `json:"views"`
	}

	stats := make([]any, 0)

	for _, stat := range resp.GetStats() {
		login, _ := selectLoginByID(stat.UserID)
		switch orderBy {
		case "likes":
			stats = append(stats, StatLikes{
				PostID:   stat.PostID,
				Login:    login,
				NumLikes: stat.Num,
			})
		case "views":
			stats = append(stats, StatViews{
				PostID:   stat.PostID,
				Login:    login,
				NumViews: stat.Num,
			})
		}
	}

	processServerResponse(c, err, stats)
}

func GetTop3Users(c *gin.Context) {
	resp, err := grpcclient.StatService.GetTop3UsersByLikes(context.Background(), &emptypb.Empty{})

	type Stat struct {
		Login    string `json:"login"`
		NumLikes uint64 `json:"likes"`
	}

	stats := make([]Stat, 0)

	for _, stat := range resp.GetStats() {
		login, _ := selectLoginByID(stat.UserID)
		stats = append(stats, Stat{
			Login:    login,
			NumLikes: stat.NumLikes,
		})
	}

	processServerResponse(c, err, stats)
}

func getAuthorID(postID uint64) (uint64, error) {
	getPostReq := pb.GetPostRequest{
		PostID: uint64(postID),
	}

	resp, err := grpcclient.PostService.GetPost(context.Background(), &getPostReq)
	if err != nil {
		return 0, err
	}

	return resp.UserID, nil
}
