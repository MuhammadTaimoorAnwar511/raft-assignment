package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/MuhammadTaimoorAnwar511/raft-assignment/kvstore"
	"github.com/MuhammadTaimoorAnwar511/raft-assignment/raft"
)

func main() {
	// 1. Parse Command-Line Flags
	idFlag := flag.String("id", "1", "Node ID (unique for each node).")
	portFlag := flag.String("port", "9001", "TCP port to listen on.")
	peersFlag := flag.String("peers", "", "Comma-separated list of peer addresses, e.g. \"localhost:9002,localhost:9003\"")
	flag.Parse()

	nodeID := *idFlag
	listenAddr := "localhost:" + *portFlag
	var peers []string
	if *peersFlag != "" {
		peers = strings.Split(*peersFlag, ",")
	}

	log.Printf("[Node %s] Starting on %s, Peers: %v\n", nodeID, listenAddr, peers)

	// 2. Create Raft Node
	raftNode, err := raft.NewRaftNode(nodeID, listenAddr, peers)
	if err != nil {
		log.Fatalf("Failed to create Raft node: %v", err)
	}

	// Start node (this sets up RPC server, election timers, heartbeat loop, etc.)
	go func() {
		if err := raftNode.Start(); err != nil {
			log.Fatalf("Error starting Raft node on %s: %v", listenAddr, err)
		}
	}()

	// 3. Key-Value Store
	store := kvstore.NewStore()

	// Let the raft node apply to the store once commands are committed
	raftNode.SetApplyCallback(store.ApplyCommand)

	// 4. Simple CLI
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("CLI ready. Enter commands (put, get, append, exit). Example: put myKey myValue")

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading line: %v", err)
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "exit":
			log.Println("Exiting CLI...")
			raftNode.Stop()
			return

		case "put":
			if len(parts) < 3 {
				fmt.Println("Usage: put <key> <value>")
				continue
			}
			key := parts[1]
			val := strings.Join(parts[2:], " ")
			err := raftNode.Propose("put", key, val)
			if err != nil {
				fmt.Printf("Error proposing put: %v\n", err)
				continue
			}
			fmt.Println("Put proposed to cluster. Waiting for commit...")

		case "append":
			if len(parts) < 3 {
				fmt.Println("Usage: append <key> <value>")
				continue
			}
			key := parts[1]
			val := strings.Join(parts[2:], " ")
			err := raftNode.Propose("append", key, val)
			if err != nil {
				fmt.Printf("Error proposing append: %v\n", err)
				continue
			}
			fmt.Println("Append proposed to cluster. Waiting for commit...")

		case "get":
			if len(parts) < 2 {
				fmt.Println("Usage: get <key>")
				continue
			}
			key := parts[1]
			value, exists := store.Get(key)
			if !exists {
				fmt.Println("(nil)")
			} else {
				fmt.Printf("%s\n", value)
			}

		default:
			fmt.Println("Unrecognized command. Available commands: put, get, append, exit")
		}
	}
}
