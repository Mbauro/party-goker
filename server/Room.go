package main

import (
	"net"
	"sync"
)

type Room struct {
	connections map[net.Conn]*Client
	mutex       sync.Mutex
	wg          sync.WaitGroup
}

type Client struct {
	isConnected bool
	hasSelected bool
	nickname    string
	choice      string
}
