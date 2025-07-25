package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"rooms/shared"
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
	firstMenu := createMessage("menu", "Welcome! Choose an option:\n1. Create a room\n2. Join an existing room")
	sendMessage(conn, firstMenu)
	// fmt.Fprintln(conn, "Welcome! Choose an option:\n1. Create a room\n2. Join an existing room")
	option, _ := reader.ReadString('\n')
	option = strings.TrimSpace(option)

	room, err := getOrCreateRoom(option, conn, reader)
	if err != nil {
		errorMsg := createMessage("error", err.Error())
		sendMessage(conn, errorMsg)
		return
	}

	defer unregisterClient(room, conn) // If the client process closes unexpectly, unregister the client and remove the connection from the room so there won't be any thread lock

	registerClient(room, conn, reader)

	for {
		mainMenu := createMessage("menu", "Choose:\n1. Vote\n2. Exit")
		sendMessage(conn, mainMenu)
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
			successMsg := createMessage("success", "Everyone has voted")
			sendMessage(conn, successMsg)
			sendVoteResults(room, conn)
			resetSelections(room)
		case "2":
			unregisterClient(room, conn)
			warningMsg := createMessage("warning", "You have left the room.")
			sendMessage(conn, warningMsg)
			return
		default:
			warningMsg := createMessage("warning", "Invalid option.")
			sendMessage(conn, warningMsg)
		}
	}
}

func sendVoteResults(room *Room, conn net.Conn) {
	var output string

	for _, conn := range room.connections {
		output += conn.nickname + " has voted " + conn.choice + "\n"
	}

	infoMsg := createMessage("info", output)
	sendMessage(conn, infoMsg)
}

func pollSelections(room *Room, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go checkMissingVoters(room, conn, ctx)

	for {
		if allClientsSelected(room) {
			break
		}
		time.Sleep(2 * time.Second)
	}
}

func checkMissingVoters(room *Room, conn net.Conn, ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var missingVoter []string

			room.mutex.Lock()
			for _, client := range room.connections {
				if !client.hasSelected {
					missingVoter = append(missingVoter, client.nickname)
				}
			}
			room.mutex.Unlock()

			if len(missingVoter) > 0 {
				warningMsg := createMessage("warning", "Missing voters: "+strings.Join(missingVoter, ","))
				sendMessage(conn, warningMsg)
			}
		}
	}
}

func getOrCreateRoom(option string, conn net.Conn, reader *bufio.Reader) (*Room, error) {

	switch option {
	case "1":
		id := uuid.New().String()
		room := &Room{connections: make(map[net.Conn]*Client), wg: sync.WaitGroup{}}
		rooms[id] = room
		successMsg := createMessage("success", "Room created. Your room UUID:"+id)
		sendMessage(conn, successMsg)
		return room, nil

	case "2":
		infoMsg := createMessage("info", "Enter your room UUID:")
		sendMessage(conn, infoMsg)
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
	infoMsg := createMessage("info", "Enter your nickname: ")
	sendMessage(conn, infoMsg)
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
		infoMsg := createMessage("info", "Select a number among: "+strings.Join(fibonacci, ", "))
		sendMessage(conn, infoMsg)
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
		errorMsg := createMessage("error", "Invalid choice.")
		sendMessage(conn, errorMsg)
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

func createMessage(msgType string, data string) shared.Message {
	return shared.Message{
		Type: msgType,
		Data: data,
	}
}

func sendMessage(conn net.Conn, message shared.Message) {
	msgBytes, _ := json.Marshal(message)

	fmt.Fprintln(conn, string(msgBytes))
}
