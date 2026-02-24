package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/you/grpc-chunk-vs-unary/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func printMem(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%s: Alloc=%d KB\n", label, m.Alloc/1024)
}

func main() {
	mode := flag.String("mode", "stream", "unary|stream")
	path := flag.String("file", "", "path to file")
	chunkSz := flag.Int("chunk", 1024*1024, "chunk size for stream")
	flag.Parse()

	if *path == "" {
		log.Fatal("missing -file")
	}

	runtime.GC()
	printMem("Start")

	conn, err := grpc.NewClient("127.0.0.1:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(1024*1024*1024),
			grpc.MaxCallRecvMsgSize(1024*1024*1024),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewFileServiceClient(conn)
	start := time.Now()

	switch *mode {
	case "unary":
		b, err := os.ReadFile(*path)
		if err != nil {
			log.Fatal(err)
		}

		printMem("After read (file in memory)")

		res, err := c.UploadUnary(context.Background(), &pb.File{Data: b})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("unary: ok=%v bytes=%d elapsed=%s\n", res.Ok, res.BytesReceived, time.Since(start))

	case "stream":
		f, err := os.Open(*path)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		printMem("After open")

		st, err := c.UploadStream(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		buf := make([]byte, *chunkSz)
		for {
			n, err := f.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			if err := st.Send(&pb.Chunk{Data: buf[:n]}); err != nil {
				log.Fatal(err)
			}
		}
		res, err := st.CloseAndRecv()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("stream: ok=%v bytes=%d elapsed=%s\n", res.Ok, res.BytesReceived, time.Since(start))

	default:
		log.Fatal("mode must be unary or stream")
	}

	printMem("Done")
}
