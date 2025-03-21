package main

import (
    "bufio"
    "flag"
    "fmt"
    "log"
    "os"
    "strings"

    // Assuming we'll import "Asg_02/raft" and "Asg_02/kvstore" 
    // Adjust the import paths to match your local structure or module name

    "github.com/MuhammadTaimoorAnwar511/raft-assignment/raft"
    "github.com/MuhammadTaimoorAnwar511/raft-assignment/kvstore"

)

func main() {
    // --- 1. Parse Command-Line Flags ---
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

    // --- 2. Initialize Raft Node ---
    // Create a new Raft Node. We'll pass in:
    // - NodeID
    // - Our own listening address
    // - Slice of peers
    // - Possibly a reference to the kvstore or pass it separately
    raftNode, err := raft.NewRaftNode(nodeID, listenAddr, peers)
    if err != nil {
        log.Fatalf("Failed to create Raft node: %v", err)
    }

    // Start listening for inbound connections (gRPC, TCP, etc.)
    // For now, let's assume the raft package handles it in `Start()`
    go func() {
        if err := raftNode.Start(); err != nil {
            log.Fatalf("Error starting Raft node on %s: %v", listenAddr, err)
        }
    }()

    // --- 3. Initialize Key-Value Store (in memory) ---
    // You may want to pass `raftNode` into the store or vice versa, 
    // depending on your design. The store is where data is actually applied once
    // commands are committed. For a placeholder:
    store := kvstore.NewStore()

    // Example: You might do something like:
    // raftNode.SetApplyCallback(store.ApplyCommand)
    // so that whenever Raft commits a log entry, it calls `store.ApplyCommand(logEntry)`

    // --- 4. Simple CLI Loop ---
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
            // Possibly instruct Raft node to gracefully stop
            raftNode.Stop()
            return

        case "put":
            // put <key> <value>
            if len(parts) < 3 {
                fmt.Println("Usage: put <key> <value>")
                continue
            }
            key := parts[1]
            val := strings.Join(parts[2:], " ")
            // Typically you'd pass the command to Raft to replicate before applying
            err := raftNode.Propose("put", key, val)
            if err != nil {
                fmt.Printf("Error proposing put: %v\n", err)
                continue
            }
            fmt.Println("Put proposed to cluster. Waiting for commit...")

        case "append":
            // append <key> <value>
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
            // get <key>
            if len(parts) < 2 {
                fmt.Println("Usage: get <key>")
                continue
            }
            key := parts[1]
            // For a GET, you might retrieve from the local kvstore
            // But remember, the store is only correct if updates have been committed
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
