package v1

import (
	"sync"
)

type Node struct {
	Name     string
	Children *[]Node
}

type Crawler struct {
	Site      string
	Stop      bool
	StopChan  chan struct{}
	Tree      Node
	URLs      chan string
	WaitGroup sync.WaitGroup
	Mutex     sync.Mutex
}
