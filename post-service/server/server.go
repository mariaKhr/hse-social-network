package server

import (
	"context"
	"time"

	"post-service/db"

	pb "post-service/protos"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PostServer struct {
	pb.UnimplementedPostServiceServer
}

func (s *PostServer) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.Post, error) {
	var postID uint64
	err := db.Pool.QueryRow(
		context.Background(),
		"INSERT INTO posts (user_id, content, created_at) VALUES ($1, $2, current_timestamp) RETURNING post_id",
		req.UserID,
		req.Content,
	).Scan(&postID)
	if err != nil {
		return nil, err
	}
	return selectPostByID(postID)
}

func (s *PostServer) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.Post, error) {
	if err := s.validateUser(req.PostID, req.UserID); err != nil {
		return nil, err
	}
	_, err := db.Pool.Exec(
		context.Background(),
		"UPDATE posts SET content = $1 WHERE post_id=$2",
		req.Content,
		req.PostID,
	)
	if err != nil {
		return nil, err
	}
	return selectPostByID(req.PostID)
}

func (s *PostServer) DeletePost(ctx context.Context, req *pb.PostCreds) (*empty.Empty, error) {
	if err := s.validateUser(req.PostID, req.UserID); err != nil {
		return nil, err
	}
	_, err := db.Pool.Exec(
		context.Background(),
		"DELETE FROM posts WHERE post_id=$1",
		req.PostID,
	)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *PostServer) GetPost(ctx context.Context, req *pb.PostCreds) (*pb.Post, error) {
	err := s.validateUser(req.PostID, req.UserID)
	if err != nil {
		return nil, err
	}
	return selectPostByID(req.PostID)
}

func (s *PostServer) GetPosts(ctx context.Context, req *pb.GetPostsRequest) (*pb.Posts, error) {
	rows, err := db.Pool.Query(
		context.Background(), `
			SELECT post_id, user_id, content, created_at FROM posts
			WHERE user_id=$1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
			`,
		req.UserID,
		req.PageSize,
		req.Page*req.PageSize)
	if err != nil {
		return nil, err
	}

	var postID, userID uint64
	var content string
	var createdAt time.Time
	var posts []*pb.Post
	_, err = pgx.ForEachRow(
		rows,
		[]any{&postID, &userID, &content, &createdAt},
		func() error {
			posts = append(posts, &pb.Post{
				PostID:    postID,
				UserID:    userID,
				Content:   content,
				CreatedAt: timestamppb.New(createdAt),
			})
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return &pb.Posts{Posts: posts}, nil
}

func (s *PostServer) validateUser(reqPostID, reqUserID uint64) error {
	var userID uint64
	err := db.Pool.QueryRow(
		context.Background(),
		"SELECT user_id FROM posts WHERE post_id=$1",
		reqPostID).Scan(&userID)

	if err != nil {
		return status.Error(codes.NotFound, "invalid post id")
	}
	if userID != reqUserID {
		return status.Error(codes.PermissionDenied, "no access to post")
	}
	return nil
}

func selectPostByID(postID uint64) (*pb.Post, error) {
	var userID uint64
	var content string
	var createdAt time.Time

	err := db.Pool.QueryRow(
		context.Background(),
		"SELECT user_id, content, created_at FROM posts WHERE post_id=$1",
		postID).Scan(&userID, &content, &createdAt)
	if err != nil {
		return nil, err
	}

	return &pb.Post{
		PostID:    postID,
		UserID:    userID,
		Content:   content,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}
