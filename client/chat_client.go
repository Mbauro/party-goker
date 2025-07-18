package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
)

func main() {
	server := flag.String("host", "", "server IP/Hostname")
	flag.Parse()

	fmt.Println(*server)

	if *server == "" {
		fmt.Println("You must specify a server to connect")
		return
	}
	conn, err := net.Dial("tcp", *server+":9000")
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()

	go func() {
		reader := bufio.NewReader(conn)
		for {
			msg, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Disconnected from server.")
				os.Exit(0)
			}
			fmt.Print(msg)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Fprintln(conn, scanner.Text())
	}
}
