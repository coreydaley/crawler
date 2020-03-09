package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/coreydaley/crawler/common"
	pb "github.com/coreydaley/crawler/crawler"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	var start = flag.String("start", "", "Start crawling the given url")
	var stop = flag.String("stop", "", "Stop crawling the given url")
	var list = flag.String("list", "", "Print a sitemap of the given url")
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewCrawlerClient(conn)

	if len(*start) != 0 {

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.Start(ctx, &pb.StartRequest{Name: *start})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("%s", r.GetMessage())
	}

	if len(*stop) != 0 {

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.Stop(ctx, &pb.StopRequest{Name: *stop})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("%s", r.GetMessage())
	}

	if len(*list) != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.List(ctx, &pb.ListRequest{Name: *list})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		ddata := common.Decompress(r.GetMessage())
		ndata := common.DecodeToNode(ddata)
		common.PrintTree(ndata, 0)
	}
}
