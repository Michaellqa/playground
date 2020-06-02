package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"kv-ttl/kv"
	"kv-ttl/pb"
	"log"
	"net"
)

type CacheServer struct {
	cache *kv.Cache
}

func (c *CacheServer) Add(ctx context.Context, r *pb.AddRequest) (*pb.Ack, error) {
	ok := c.cache.Add(r.Key, kv.T{V: r.Value.Value})
	return &pb.Ack{Ok: ok}, nil
}

func (c *CacheServer) Get(ctx context.Context, r *pb.GetRequest) (*pb.GetResponse, error) {
	value, ok := c.cache.Get(r.Key)
	return &pb.GetResponse{Ok: ok, Value: &pb.T{Value: value.V}}, nil
}

//	GetAll(ctx context.Context, in *Empty, opts ...grpc.CallOption) (Storage_GetAllClient, error)

func (c *CacheServer) GetAll(_ *pb.Empty, stream pb.Storage_GetAllServer) error {
	for _, v := range c.cache.GetAll() {
		if err := stream.Send(&pb.T{Value: v.V}); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}

	opts := []grpc.ServerOption{
		//grpc.KeepaliveParams(keepalive.ServerParameters{
		//	MaxConnectionIdle: 5 * time.Minute,
		//}),
	}

	grpcServer := grpc.NewServer(opts...)

	cacheServer := &CacheServer{cache: kv.NewCache()}
	pb.RegisterStorageServer(grpcServer, cacheServer)
	grpcServer.Serve(listener)
}
