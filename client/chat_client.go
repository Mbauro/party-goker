package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:9000")
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
