package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"

	"github.com/you/grpc-chunk-vs-unary/pb"
	"google.golang.org/grpc"
)

type svc struct {
	pb.UnimplementedFileServiceServer
}

func (s *svc) UploadUnary(ctx context.Context, req *pb.File) (*pb.UploadStatus, error) {
	if err := os.WriteFile("out_unary.bin", req.Data, 0644); err != nil {
		return nil, err
	}
	return &pb.UploadStatus{Ok: true, BytesReceived: int64(len(req.Data))}, nil
}

func (s *svc) UploadStream(stream pb.FileService_UploadStreamServer) error {
	f, err := os.Create("out_stream.bin")
	if err != nil {
		return err
	}
	defer f.Close()

	var total int64
	for {
		ch, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.UploadStatus{Ok: true, BytesReceived: total})
		}
		if err != nil {
			return err
		}
		n, err := f.Write(ch.Data)
		if err != nil {
			return err
		}
		total += int64(n)
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024 * 1024 * 1024),
	)
	pb.RegisterFileServiceServer(s, &svc{})

	log.Println("server listening :50051")
	log.Fatal(s.Serve(lis))
}
