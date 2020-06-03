package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"kv-ttl/kv"
	"kv-ttl/pb"
	"log"
	"net"
)

const (
	notFound  = "not_found"
	duplicate = "duplicate"
)

type CacheServer struct {
	cache *kv.Cache
}

func (c *CacheServer) Add(ctx context.Context, r *pb.KeyValue) (*pb.Empty, error) {
	ok := c.cache.Add(r.Key, kv.T{V: r.Value.Value})
	if ok {
		return &pb.Empty{}, fmt.Errorf(duplicate)
	}
	return &pb.Empty{}, nil
}

func (c *CacheServer) Get(ctx context.Context, r *pb.Key) (*pb.T, error) {
	value, ok := c.cache.Get(r.Key)
	if !ok {
		return &pb.T{Value: value.V}, fmt.Errorf(notFound)
	}
	return &pb.T{Value: value.V}, nil
}

func (c *CacheServer) GetAll(req *pb.Empty, stream pb.Storage_GetAllServer) error {
	for _, v := range c.cache.GetAll() {
		if err := stream.Send(&pb.T{Value: v.V}); err != nil {
			return err
		}
	}
	return nil
}

func (c *CacheServer) Remove(ctx context.Context, req *pb.Key) (*pb.Empty, error) {
	c.cache.Remove(req.Key)
	return nil, nil
}
func (c *CacheServer) AddWithTtl(ctx context.Context, req *pb.KeyValueTtl) (*pb.Empty, error) {
	dur, err := ptypes.Duration(req.Ttl)
	if err != nil {
		return nil, err
	}
	ok := c.cache.AddWithTtl(req.Key, kv.T{V: req.Value.Value}, dur)
	if !ok {
		return &pb.Empty{}, fmt.Errorf(duplicate)
	}
	return &pb.Empty{}, nil
}
func (c *CacheServer) GetTtl(ctx context.Context, req *pb.Key) (*pb.TtlResponse, error) {
	dur, ok := c.cache.GetTtl(req.Key)
	if !ok {
		return &pb.TtlResponse{}, fmt.Errorf(notFound)
	}
	return &pb.TtlResponse{Ttl: ptypes.DurationProto(dur)}, nil
}
func (c *CacheServer) SetTtl(ctx context.Context, req *pb.TtlRequest) (*pb.Empty, error) {
	t, err := ptypes.Timestamp(req.Stamp)
	if err != nil {
		return &pb.Empty{}, err
	}
	ok := c.cache.SetTtl(req.Key, &t)
	if ok {
		return &pb.Empty{}, fmt.Errorf(notFound)
	}
	return &pb.Empty{}, nil
}

func main() {
	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}

	opts := []grpc.ServerOption{}

	grpcServer := grpc.NewServer(opts...)

	cacheServer := &CacheServer{cache: kv.NewCache()}
	pb.RegisterStorageServer(grpcServer, cacheServer)
	grpcServer.Serve(listener)
}
