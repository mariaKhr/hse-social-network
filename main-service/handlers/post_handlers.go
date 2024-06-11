package handlers

import (
	"context"
	"net/http"
	"strconv"

	grpcclient "main-service/grpc_client"
	pb "main-service/protos"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreatePost(c *gin.Context) {
	createPostReq := pb.CreatePostRequest{}
	if err := c.BindJSON(&createPostReq); err != nil {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(err.Error())
		return
	}

	createPostReq.UserID = getUserID(c)

	resp, err := grpcclient.PostService.CreatePost(context.Background(), &createPostReq)
	processServerResponse(c, err, resp)
}

func UpdatePost(c *gin.Context) {
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(err.Error())
		return
	}

	updatePostReq := pb.UpdatePostRequest{
		UserID: getUserID(c),
		PostID: uint64(postId),
	}
	if err := c.BindJSON(&updatePostReq); err != nil {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(err.Error())
		return
	}

	updatePostReq.UserID = getUserID(c)

	resp, err := grpcclient.PostService.UpdatePost(context.Background(), &updatePostReq)
	processServerResponse(c, err, resp)
}

func DeletePost(c *gin.Context) {
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(err.Error())
		return
	}

	deletePostReq := pb.DeletePostRequest{
		UserID: getUserID(c),
		PostID: uint64(postId),
	}

	resp, err := grpcclient.PostService.DeletePost(context.Background(), &deletePostReq)
	processServerResponse(c, err, resp)
}

func GetPost(c *gin.Context) {
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(err.Error())
		return
	}

	getPostReq := pb.GetPostRequest{
		PostID: uint64(postId),
	}

	resp, err := grpcclient.PostService.GetPost(context.Background(), &getPostReq)
	processServerResponse(c, err, resp)
}

func GetPage(c *gin.Context) {
	pageQuery, ok := c.GetQuery("page")
	if !ok {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString("no page in query")
		return
	}

	pageSizeQuery, ok := c.GetQuery("pageSize")
	if !ok {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString("no pageSize in query")
		return
	}

	userIDQuery, ok := c.GetQuery("userId")
	if !ok {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString("no userId in query")
		return
	}

	page, err := strconv.Atoi(pageQuery)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(err.Error())
		return
	}

	pageSize, err := strconv.Atoi(pageSizeQuery)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(err.Error())
		return
	}

	userID, err := strconv.Atoi(userIDQuery)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(err.Error())
		return
	}

	getPostsReq := pb.GetPostsRequest{
		UserID:   uint64(userID),
		Page:     uint32(page),
		PageSize: uint32(pageSize),
	}

	resp, err := grpcclient.PostService.GetPosts(context.Background(), &getPostsReq)
	processServerResponse(c, err, resp)
}

func processServerResponse(c *gin.Context, err error, resp any) {
	if status.Code(err) == codes.NotFound {
		c.Status(http.StatusNotFound)
		c.Writer.WriteString(err.Error())
		return
	}
	if status.Code(err) == codes.PermissionDenied {
		c.Status(http.StatusForbidden)
		c.Writer.WriteString(err.Error())
		return
	}
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Writer.WriteString(err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

func getUserID(c *gin.Context) uint64 {
	strUserID, _ := c.Get("userID")
	return uint64(strUserID.(float64))
}
