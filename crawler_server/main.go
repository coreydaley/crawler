package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreydaley/crawler/common"
	pb "github.com/coreydaley/crawler/crawler"
	"github.com/gocolly/colly/v2"
	"google.golang.org/grpc"

	v1 "github.com/coreydaley/crawler/api/v1"
)

const (
	port = ":50051"
)

var (
	crawlers     = map[string]*v1.Crawler{}
	goroutines   = 100 // runtime.NumCPU() could be used here also
	channelCache = 100 // a larger number here allows the crawling to happen faster
)

func BuildTree(n int, crawler *v1.Crawler) {
	defer crawler.WaitGroup.Done()
	for u := range crawler.URLs {
		select {
		case <-crawler.StopChan:
			break
		default:
			u, err := url.Parse(u)
			if err != nil {
				fmt.Println(fmt.Sprintf("%#v", err))
			}
			fmt.Println(fmt.Sprintf("Go Routine %d: Processing %s", n, u.Path))
			AddToTree(crawler, &crawler.Tree, strings.Split(strings.Trim(u.Path, "/"), "/"))
		}

	}
	fmt.Println(fmt.Sprintf("Go Routine %d: Stopping Processing", n))

}

func AddToTree(crawler *v1.Crawler, root *v1.Node, parts []string) {
	if len(parts) == 0 {
		return
	}
	fmt.Println(fmt.Sprintf("Adding %s to site tree", strings.Join(parts, "/")))
	name := parts[0]
	if len(parts) > 1 {
		parts = parts[1:]
	} else {
		parts = []string{}
	}

	for _, c := range *root.Children {
		if c.Name == name {
			AddToTree(crawler, &c, parts)
			return
		}
	}
	if len(name) == 0 {
		return
	}
	node := v1.Node{Name: name, Children: &[]v1.Node{}}
	crawler.Mutex.Lock()
	*root.Children = append(*root.Children, node)
	crawler.Mutex.Unlock()
	AddToTree(crawler, &node, parts)

}

func Process(crawler *v1.Crawler, root string) {

	start := time.Now()
	for i := 0; i < goroutines-1; i++ {
		crawler.WaitGroup.Add(1)
		go BuildTree(i, crawler)
	}

	crawler.Tree = v1.Node{Name: root, Children: &[]v1.Node{}}

	c := colly.NewCollector(
		colly.AllowedDomains(root),
		colly.Async(true),
	)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: goroutines})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if !crawler.Stop {
			c.Visit(e.Request.AbsoluteURL(link))
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
		crawler.URLs <- r.URL.String()
	})

	c.Visit(fmt.Sprintf("https://%s/", root))

	c.Wait()
	close(crawler.URLs)
	crawler.WaitGroup.Wait()
	fmt.Println(fmt.Sprintf("Processed %s in %v", crawler.Site, time.Since(start)))
}

type server struct {
	pb.UnimplementedCrawlerServer
}

func (s *server) List(ctx context.Context, in *pb.ListRequest) (*pb.ListReply, error) {
	if crawler, ok := crawlers[in.GetName()]; ok {
		data := common.EncodeToBytes(crawler.Tree)
		cdata := common.Compress(data)
		return &pb.ListReply{Message: cdata}, nil
	}
	return &pb.ListReply{Message: nil}, nil
}

func (s *server) Start(ctx context.Context, in *pb.StartRequest) (*pb.StartReply, error) {
	u, err := url.Parse(in.GetName())
	if err != nil {
		fmt.Println(fmt.Sprintf("%#v", err))
	}
	crawler := v1.Crawler{
		Site:      u.Path,
		Stop:      false,
		StopChan:  make(chan struct{}),
		Tree:      v1.Node{},
		URLs:      make(chan string, channelCache),
		WaitGroup: sync.WaitGroup{},
		Mutex:     sync.Mutex{},
	}
	crawlers[u.Path] = &crawler
	go Process(&crawler, u.Path)

	return &pb.StartReply{Message: "Started Crawler " + in.GetName()}, nil
}

func (s *server) Stop(ctx context.Context, in *pb.StopRequest) (*pb.StopReply, error) {
	if crawler, ok := crawlers[in.GetName()]; ok {
		crawler.Stop = true
		close(crawler.StopChan)
	}

	return &pb.StopReply{Message: "Stopped Crawler " + in.GetName()}, nil
}

func main() {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCrawlerServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
