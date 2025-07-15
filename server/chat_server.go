package main

import (
	"bufio"
	"fmt"
	"time"

	"net"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"
)

var (
	roomMutex sync.Mutex
	rooms     = make(map[string]*Room) // UUID -> Room
)

func main() {
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Chat server listening on port 9000...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {

	defer conn.Close()

	reader := bufio.NewReader(conn)
	fmt.Fprintln(conn, "Welcome! Choose an option:\n1. Create a room\n2. Join an existing room")
	option, _ := reader.ReadString('\n')
	option = strings.TrimSpace(option)

	room, err := getOrCreateRoom(option, conn, reader)
	if err != nil {
		fmt.Fprintln(conn, err)
		return
	}

	defer unregisterClient(room, conn) // If the client process closes unexpectly, unregister the client and remove the connection from the room so there won't be any thread lock

	registerClient(room, conn, reader)

	for {
		fmt.Fprintln(conn, "Choose:\n1. Vote\n2. Exit")
		choice, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		switch strings.TrimSpace(choice) {
		case "1":
			if err := handleVoting(room, conn, reader); err != nil {
				return
			}
			room.wg.Add(1)
			go pollSelections(room, conn, &room.wg)
			room.wg.Wait()
			fmt.Fprintln(conn, "Everyone has voted")
			sendVoteResults(room, conn)
			resetSelections(room)
		case "2":
			unregisterClient(room, conn)
			fmt.Fprintln(conn, "You have left the room.")
			return
		default:
			fmt.Fprintln(conn, "Invalid option.")
		}
	}
}

func sendVoteResults(room *Room, conn net.Conn) {
	var output string

	for _, conn := range room.connections {
		output += conn.nickname + " has voted " + conn.choice + "\n"
	}

	fmt.Fprintln(conn, output)
}

func pollSelections(room *Room, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		if allClientsSelected(room) {
			break
		}
		time.Sleep(2 * time.Second)
	}
}

func getOrCreateRoom(option string, conn net.Conn, reader *bufio.Reader) (*Room, error) {

	switch option {
	case "1":
		id := uuid.New().String()
		room := &Room{connections: make(map[net.Conn]*Client), wg: sync.WaitGroup{}}
		rooms[id] = room
		fmt.Fprintln(conn, "Room created. Your room UUID:", id)
		return room, nil

	case "2":
		fmt.Fprintln(conn, "Enter your room UUID:")
		roomID, _ := reader.ReadString('\n')
		roomID = strings.TrimSpace(roomID)

		roomMutex.Lock()
		defer roomMutex.Unlock()
		room, exists := rooms[roomID]

		if !exists {
			return nil, fmt.Errorf("Room not found")
		}

		return room, nil

	default:
		return nil, fmt.Errorf("Invalid option")
	}
}

func registerClient(room *Room, conn net.Conn, reader *bufio.Reader) {
	fmt.Fprintln(conn, "Enter your nickname: ")
	nickname, _ := reader.ReadString('\n')
	nickname = strings.TrimSpace(nickname)
	room.connections[conn] = &Client{isConnected: true, nickname: nickname}
}

func unregisterClient(room *Room, conn net.Conn) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	delete(room.connections, conn)
}

func handleVoting(room *Room, conn net.Conn, reader *bufio.Reader) error {

	fibonacci := []string{"0", "1", "2", "3", "5", "8", "13", "21", "34", "55", "89"}

	var vote string
	for {
		fmt.Fprintln(conn, "Select a number among: "+strings.Join(fibonacci, ", "))
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		vote = strings.TrimSpace(input)

		if slices.Contains(fibonacci, vote) {
			setClientSelected(room, conn)
			room.connections[conn].choice = vote
			break
		}
		fmt.Fprintln(conn, "Invalid choice.")
	}

	return nil
}

func resetSelections(room *Room) {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	for _, conn := range room.connections {
		conn.hasSelected = false
	}
}

func setClientSelected(room *Room, conn net.Conn) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if client, ok := room.connections[conn]; ok {
		client.hasSelected = true
	}
}

func allClientsSelected(room *Room) bool {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	for _, client := range room.connections {
		if !client.hasSelected {
			return false
		}
	}
	return true
}

func broadcast(room *Room, sender net.Conn, message string) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	for conn := range room.connections {
		if conn != sender {
			fmt.Fprintln(conn, message)
		}
	}
}
