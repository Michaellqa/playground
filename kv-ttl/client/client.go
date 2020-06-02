package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"kv-ttl/kv"
	"kv-ttl/pb"
	"log"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	cl := pb.NewStorageClient(conn)

	ctx := context.Background()
	_, err = cl.Add(ctx, &pb.AddRequest{Key: "1", Value: &pb.T{Value: "One"}})
	_, err = cl.Add(ctx, &pb.AddRequest{Key: "2", Value: &pb.T{Value: "Two"}})
	_, err = cl.Add(ctx, &pb.AddRequest{Key: "3", Value: &pb.T{Value: "Three"}})

	resp, err := cl.Get(ctx, &pb.GetRequest{Key: "0"})
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("#0: %v, %v\n", resp.Ok, resp.Value)

	resp, err = cl.Get(ctx, &pb.GetRequest{Key: "1"})
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("#1: %v, %v\n", resp.Ok, resp.Value)

	stream, err := cl.GetAll(ctx, &pb.Empty{})
	var values []kv.T
	for {
		value, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		values = append(values, kv.T{V: value.Value})
	}
	fmt.Printf("#all: %v", values)
}
