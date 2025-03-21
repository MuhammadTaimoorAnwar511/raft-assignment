package raft

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/MuhammadTaimoorAnwar511/raft-assignment/internal"
)

// RaftNode holds state for one Raft server.
type RaftNode struct {
	// Identifiers
	ID      string
	Address string   // e.g. "localhost:9001"
	Peers   []string // addresses of peers

	// Persistent State
	currentTerm int
	votedFor    string
	log         []internal.LogEntry

	// Volatile State
	commitIndex int
	lastApplied int

	// Volatile leader-only state
	nextIndex  map[string]int
	matchIndex map[string]int

	// Raft node state
	state internal.RaftState

	// Concurrency
	mu      sync.Mutex
	applyCh chan internal.LogEntry
	stopCh  chan struct{}
	wg      sync.WaitGroup

	applyCallback      func(internal.Command)
	electionResetEvent time.Time
	listener           net.Listener // for RPC
}

// NewRaftNode initializes a Raft node.
func NewRaftNode(id, address string, peers []string) (*RaftNode, error) {
	node := &RaftNode{
		ID:                 id,
		Address:            address,
		Peers:              peers,
		currentTerm:        0,
		votedFor:           "",
		log:                make([]internal.LogEntry, 0),
		commitIndex:        0,
		lastApplied:        0,
		nextIndex:          make(map[string]int),
		matchIndex:         make(map[string]int),
		state:              internal.Follower,
		applyCh:            make(chan internal.LogEntry, 100),
		stopCh:             make(chan struct{}),
		electionResetEvent: time.Now(),
	}
	return node, nil
}

// Start begins the Raft node's operation.
func (rn *RaftNode) Start() error {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	fmt.Printf("[Node %s] Starting. State = Follower\n", rn.ID)

	// Start RPC server
	if err := rn.startRPCServer(); err != nil {
		return err
	}

	// Start background goroutines
	rn.wg.Add(2)
	go rn.runElectionTimer()
	go rn.runHeartbeatLoop()

	return nil
}

// Stop shuts down the Raft node.
func (rn *RaftNode) Stop() {
	close(rn.stopCh)
	rn.listener.Close()
	rn.wg.Wait()
	fmt.Printf("[Node %s] Stopped.\n", rn.ID)
}

// Propose handles client proposals.
func (rn *RaftNode) Propose(cmdType string, key string, value string) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.state != internal.Leader {
		return fmt.Errorf("Node %s is not the leader. Propose ignored.", rn.ID)
	}

	command := internal.Command{
		Key:   key,
		Value: value,
	}
	switch cmdType {
	case "put":
		command.Type = internal.CommandPut
	case "append":
		command.Type = internal.CommandAppend
	case "get":
		command.Type = internal.CommandGet
	default:
		return fmt.Errorf("unknown command type %s", cmdType)
	}

	fmt.Printf("[Node %s][Leader] Proposing command: %+v\n", rn.ID, command)
	return nil
}

// SetApplyCallback sets a function for committed log entries.
func (rn *RaftNode) SetApplyCallback(cb func(internal.Command)) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.applyCallback = cb
}

// lastLogIndex returns the last log index.
func (rn *RaftNode) lastLogIndex() int {
	if len(rn.log) == 0 {
		return 0
	}
	return rn.log[len(rn.log)-1].Index
}

// lastLogTerm returns the last log term.
func (rn *RaftNode) lastLogTerm() int {
	if len(rn.log) == 0 {
		return 0
	}
	return rn.log[len(rn.log)-1].Term
}
