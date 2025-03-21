// raft.go - Final Cleaned Version (lastLogIndex and lastLogTerm centralized here)
package raft

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/MuhammadTaimoorAnwar511/raft-assignment/internal"
)

type RaftNode struct {
	ID      string
	Address string
	Peers   []string

	currentTerm int
	votedFor    string
	log         []internal.LogEntry

	commitIndex int
	lastApplied int

	nextIndex  map[string]int
	matchIndex map[string]int

	state internal.RaftState

	mu      sync.Mutex
	applyCh chan internal.LogEntry
	stopCh  chan struct{}
	wg      sync.WaitGroup

	applyCallback      func(internal.Command)
	electionResetEvent time.Time
	listener           net.Listener
}

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

func (rn *RaftNode) Start() error {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	fmt.Printf("[Node %s] Starting. State = Follower\n", rn.ID)
	if err := rn.startRPCServer(); err != nil {
		return err
	}
	rn.wg.Add(2)
	go rn.runElectionTimer()
	go rn.runHeartbeatLoop()
	return nil
}

func (rn *RaftNode) Stop() {
	close(rn.stopCh)
	rn.listener.Close()
	rn.wg.Wait()
	fmt.Printf("[Node %s] Stopped.\n", rn.ID)
}

func (rn *RaftNode) Propose(cmdType string, key string, value string) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	if rn.state != internal.Leader {
		return fmt.Errorf("Node %s is not the leader. Propose ignored.", rn.ID)
	}

	command := internal.Command{Key: key, Value: value}
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

func (rn *RaftNode) SetApplyCallback(cb func(internal.Command)) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.applyCallback = cb
}

// Centralized lastLogIndex and lastLogTerm helpers
func (rn *RaftNode) LastLogIndex() int {
	if len(rn.log) == 0 {
		return 0
	}
	return rn.log[len(rn.log)-1].Index
}

func (rn *RaftNode) LastLogTerm() int {
	if len(rn.log) == 0 {
		return 0
	}
	return rn.log[len(rn.log)-1].Term
}
