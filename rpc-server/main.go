package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"

	pb "ttti-2023-asgn/rpc-server/proto"

	"github.com/redis/go-redis/v9"
)

const (
	port = ":8888"
)

var (
	rdb = &RedisClient{}
)

type RedisClient struct {
	cli *redis.Client
}

func (c *RedisClient) InitClient(ctx context.Context, address, password string) error {
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	// test connection
	if err := r.Ping(ctx).Err(); err != nil {
		return err
	}

	c.cli = r
	return nil
}

type messageServer struct {
	pb.UnimplementedMessageServiceServer
}

func (s *messageServer) Send(ctx context.Context, req *pb.SendRequest) (*pb.SendResponse, error) {
	if err := validateSendRequest(req); err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	message := &Message{
		Message:   req.Text,
		Sender:    req.Sender,
		Timestamp: timestamp,
	}
	roomID, err := getRoomID(req.Chat)
	if err != nil {
		return nil, err
	}
	err = rdb.SaveMessage(ctx, roomID, message)
	if err != nil {
		return nil, err
	}
	resp := &pb.SendResponse{}
	return resp, nil
}

func (s *messageServer) Pull(ctx context.Context, req *pb.PullRequest) (*pb.PullResponse, error) {
	if req.Cursor == 0 {
		req.Cursor = 0
		fmt.Println("set cursor to 0")
	}

	roomID, err := getRoomID(req.Chat)
	if err != nil {
		return nil, err
	}

	start := req.Cursor

	end := start + int64(req.Limit) // did not -1 for hasMore check

	messages, err := rdb.GetMessagesByRoomID(ctx, roomID, start, end, req.Reverse)
	if err != nil {
		return nil, err
	}

	respMessages := make([]*pb.Message, 0)
	var counter int32 = 0
	var nextCursor int64 = 0
	hasMore := false
	for _, msg := range messages {
		if counter+1 > req.Limit {
			// having extra value here means it has more data
			hasMore = true
			nextCursor = end
			break // do not return the last message
		}
		temp := &pb.Message{
			Chat:     req.Chat,
			Text:     msg.Message,
			Sender:   msg.Sender,
			SendTime: msg.Timestamp,
		}
		respMessages = append(respMessages, temp)
		counter += 1
	}

	resp := &pb.PullResponse{}
	resp.Messages = respMessages
	resp.HasMore = hasMore
	resp.NextCursor = nextCursor

	return resp, nil
}

func getRoomID(chat string) (string, error) {
	var roomID string

	lowercase := strings.ToLower(chat)
	senders := strings.Split(lowercase, ":")
	if len(senders) != 2 {
		err := fmt.Errorf("invalid Chat ID '%s', should be in the format of user1:user2", chat)
		return "", err
	}

	sender1, sender2 := senders[0], senders[1]
	// Compare the sender and receiver alphabetically, and sort them asc to form the room ID
	if comp := strings.Compare(sender1, sender2); comp == 1 {
		roomID = fmt.Sprintf("%s:%s", sender2, sender1)
	} else {
		roomID = fmt.Sprintf("%s:%s", sender1, sender2)
	}

	return roomID, nil
}

func validateSendRequest(req *pb.SendRequest) error {
	senders := strings.Split(req.Chat, ":")
	if len(senders) != 2 {
		err := fmt.Errorf("invalid Chat ID '%s', should be in the format of user1:user2", req.Chat)
		return err
	}
	sender1, sender2 := senders[0], senders[1]

	if req.Sender != sender1 && req.Sender != sender2 {
		err := fmt.Errorf("sender '%s' not in the chat room", req.Sender)
		return err
	}

	return nil
}

type Message struct {
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

func (c *RedisClient) SaveMessage(ctx context.Context, roomID string, message *Message) error {
	// Marshal the Go struct into JSON bytes
	text, err := json.Marshal(message)
	if err != nil {
		return err
	}

	member := &redis.Z{
		Score:  float64(message.Timestamp), // The sort key
		Member: text,                       // Data
	}

	_, err = c.cli.ZAdd(ctx, roomID, *member).Result()
	if err != nil {
		return err
	}

	return nil
}

func (c *RedisClient) GetMessagesByRoomID(ctx context.Context, roomID string, start, end int64, reverse bool) ([]*Message, error) {
	var (
		rawMessages []string
		messages    []*Message
		err         error
	)

	if reverse {
		// Desc order with time -> first message is the latest message
		rawMessages, err = c.cli.ZRevRange(ctx, roomID, start, end).Result()
		if err != nil {
			return nil, err
		}
	} else {
		// Asc order with time -> first message is the earliest message
		rawMessages, err = c.cli.ZRange(ctx, roomID, start, end).Result()
		if err != nil {
			return nil, err
		}
	}

	for _, msg := range rawMessages {
		temp := &Message{}
		err := json.Unmarshal([]byte(msg), temp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, temp)
	}

	return messages, nil
}

func main() {
	ctx := context.Background()
	// Initialize Redis client
	err := rdb.InitClient(ctx, "redis:6379", "")
	if err != nil {
		errMsg := fmt.Sprintf("failed to init Redis client, err: %v", err)
		log.Fatal(errMsg)
	}

	listen, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMessageServiceServer(grpcServer, &messageServer{})

	fmt.Println("gRPC server started on port " + port)
	if err := grpcServer.Serve(listen); err != nil {
		panic(err)
	}
}
