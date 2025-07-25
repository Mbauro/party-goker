package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"rooms/shared"

	"github.com/fatih/color"
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
			jsonMsg := shared.Message{}
			json.Unmarshal([]byte(msg), &jsonMsg)
			printMessage(jsonMsg)
			//fmt.Print(msg)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Fprintln(conn, scanner.Text())
	}
}

func printMessage(jsonMsg shared.Message) {
	switch jsonMsg.Type {
	case "menu":
		color.HiCyan(jsonMsg.Data)
	case "error":
		color.HiRed(jsonMsg.Data)
	case "info":
		color.HiMagenta(jsonMsg.Data)
	case "success":
		color.HiGreen(jsonMsg.Data)
	case "warning":
		color.HiYellow(jsonMsg.Data)
	default:
		color.White(jsonMsg.Data)
	}

}
