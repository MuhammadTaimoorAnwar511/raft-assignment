package raft

import (
    "sync"
    "time"
)

type ServerState int

const (
    Follower ServerState = iota
    Candidate
    Leader
)

// LogEntry holds a single command and the term when it was created.
type LogEntry struct {
    Term    int
    Command interface{}
}

// RaftNode is the core struct holding Raft state.
type RaftNode struct {
    mu          sync.Mutex
    ID          int           // identifier for this node
    State       ServerState   // follower, candidate, or leader
    CurrentTerm int           // latest term server has seen
    VotedFor    int           // candidateID that received vote in current term
    Log         []LogEntry    // log entries

    CommitIndex int
    LastApplied int

    NextIndex  []int // for each follower
    MatchIndex []int // for each follower

    Peers []string // addresses of other nodes

    // Timeouts
    ElectionTimeout  time.Duration
    HeartbeatTimeout time.Duration

    // Channels or other communication means
    // e.g., MessageCh chan Message
}

// NewRaftNode is a helper to create a RaftNode with default config
func NewRaftNode(id int, peers []string) *RaftNode {
    return &RaftNode{
        ID:              id,
        State:           Follower,
        CurrentTerm:     0,
        VotedFor:        -1, // -1 indicates no vote
        Log:             make([]LogEntry, 0),
        CommitIndex:     0,
        LastApplied:     0,
        NextIndex:       make([]int, len(peers)),
        MatchIndex:      make([]int, len(peers)),
        Peers:           peers,
        ElectionTimeout:  time.Millisecond * 300,  // example
        HeartbeatTimeout: time.Millisecond * 100,  // example
    }
}
