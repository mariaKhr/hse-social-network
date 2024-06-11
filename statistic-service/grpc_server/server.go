package grpcserver

import (
	"context"
	"database/sql"
	"fmt"
	"statistic-service/db"
	pb "statistic-service/protos"

	"google.golang.org/protobuf/types/known/emptypb"
)

type StatServer struct {
	pb.UnimplementedStatisticServiceServer
}

func (*StatServer) GetStatistics(ctx context.Context, req *pb.GetStatisticsRequest) (*pb.GetStatisticsResponse, error) {
	var authorID uint64
	var numLikes uint64
	db.Conn.QueryRow(`
		SELECT author_id, countDistinct(user_id)
		FROM likes
		WHERE post_id = $1
		GROUP BY post_id, author_id
	`, req.PostID).Scan(&authorID, &numLikes)

	var numViews uint64
	db.Conn.QueryRow(`
		SELECT countDistinct(user_id)
		FROM views
		WHERE post_id = $1
		GROUP BY post_id, author_id
	`, req.PostID).Scan(&numViews)

	return &pb.GetStatisticsResponse{
		PostID:   req.PostID,
		UserID:   authorID,
		NumLikes: numLikes,
		NumViews: numViews,
	}, nil
}

func (*StatServer) GetTop5Posts(ctx context.Context, req *pb.GetTop5PostsRequest) (*pb.GetTop5PostsResponse, error) {
	var tableName string
	switch *req.Comparator.Enum() {
	case pb.Comparator_LIKES:
		tableName = "likes"
	case pb.Comparator_VIEWS:
		tableName = "views"
	}

	rows, err := db.Conn.Query(fmt.Sprintf(`
		SELECT author_id, post_id, countDistinct(user_id) as count
		FROM %v
		GROUP BY post_id, author_id
		ORDER BY count DESC
		LIMIT 5
	`, tableName),
	)
	if err == sql.ErrNoRows {
		return &pb.GetTop5PostsResponse{}, nil
	}
	if err != nil {
		return nil, err
	}

	var posts []*pb.GetTop5PostsResponse_Stat
	for rows.Next() {
		var authorID, postID, count uint64
		if err := rows.Scan(&authorID, &postID, &count); err != nil {
			return nil, err
		}
		posts = append(posts, &pb.GetTop5PostsResponse_Stat{
			PostID: postID,
			UserID: authorID,
			Num:    count,
		})
	}

	return &pb.GetTop5PostsResponse{Stats: posts}, nil
}

func (*StatServer) GetTop3UsersByLikes(ctx context.Context, req *emptypb.Empty) (*pb.GetTop3UsersByLikesResponse, error) {
	rows, err := db.Conn.Query(`
		SELECT author_id, sum(likes_count) as total_likes
		FROM (
			SELECT author_id, post_id, countDistinct(user_id) as likes_count
			FROM likes
			GROUP BY post_id, author_id
		)
		GROUP BY author_id
		ORDER BY total_likes DESC
		LIMIT 3
	`,
	)
	if err == sql.ErrNoRows {
		return &pb.GetTop3UsersByLikesResponse{}, nil
	}
	if err != nil {
		return nil, err
	}

	var users []*pb.GetTop3UsersByLikesResponse_Stat
	for rows.Next() {
		var userID, count uint64
		if err := rows.Scan(&userID, &count); err != nil {
			return nil, err
		}
		users = append(users, &pb.GetTop3UsersByLikesResponse_Stat{
			UserID:   userID,
			NumLikes: count,
		})
	}

	return &pb.GetTop3UsersByLikesResponse{Stats: users}, nil
}
