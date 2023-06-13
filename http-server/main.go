package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"ttti-2023-asgn/http-server/proto"

	"google.golang.org/grpc"
)

var (
	port        = ":8080"
	grpcAddress = "rpc-server:8888"
)

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	http.HandleFunc("/api/send", sendMessageHandler)
	http.HandleFunc("/api/pull", pullMessageHandler)

	println("HTTP server started on port " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req proto.SendRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	conn, err := grpc.Dial(grpcAddress, grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to RPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := proto.NewMessageServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resp, err := client.Send(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to send gRPC request", http.StatusInternalServerError)
		return
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal gRPC response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}

func pullMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Define a struct to hold the request body parameters
	type PullRequestParams struct {
		Chat    string `json:"chat"`
		Cursor  *int64 `json:"cursor"`
		Limit   *int32 `json:"limit"`
		Reverse bool   `json:"reverse"`
	}

	// Define a default value for the Cursor field
	defaultCursor := int64(0)
	defaultLimit := int32(10)

	// Define a struct with default values to unmarshal the request body
	var defaultParams = PullRequestParams{
		Cursor: &defaultCursor,
		Limit:  &defaultLimit,
	}

	// Unmarshal the request body into the struct
	var params PullRequestParams
	err = json.Unmarshal(body, &params)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// If the request body does not contain a Cursor field, set it to the default value
	if params.Cursor == nil {
		params.Cursor = defaultParams.Cursor
	}

	// If the request body does not contain a Limit field, set it to the default value
	if params.Limit == nil {
		params.Limit = defaultParams.Limit
	}

	// Create the gRPC client connection
	conn, err := grpc.Dial(grpcAddress, grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to RPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Create the gRPC client
	client := proto.NewMessageServiceClient(conn)

	// Create the gRPC request
	req := &proto.PullRequest{
		Chat:    params.Chat,
		Cursor:  *params.Cursor,
		Limit:   *params.Limit,
		Reverse: params.Reverse,
	}

	// Invoke the gRPC method
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resp, err := client.Pull(ctx, req)
	if err != nil {
		http.Error(w, "Failed to send gRPC request", http.StatusInternalServerError)
		return
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal gRPC response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}
