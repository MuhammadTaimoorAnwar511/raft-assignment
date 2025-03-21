package raft

import (
    "fmt"
    "time"
)

// Example: run a goroutine that tracks election timeouts
func (rn *RaftNode) runElectionTimer() {
    // This is just a very minimal skeleton
    go func() {
        for {
            time.Sleep(rn.ElectionTimeout)
            rn.mu.Lock()
            if rn.State == Follower || rn.State == Candidate {
                // Start election
                rn.startElection()
            }
            rn.mu.Unlock()
        }
    }()
}

func (rn *RaftNode) startElection() {
    rn.State = Candidate
    rn.CurrentTerm++
    rn.VotedFor = rn.ID
    votes := 1 // vote for self
    fmt.Printf("[Node %d] Starting election for term %d\n", rn.ID, rn.CurrentTerm)

    // Pseudo: send RequestVote RPCs to peers
    // If majority votes -> becomeLeader()
}

func (rn *RaftNode) becomeLeader() {
    rn.State = Leader
    fmt.Printf("[Node %d] Became Leader in term %d\n", rn.ID, rn.CurrentTerm)

    // Initialize nextIndex and matchIndex
    for i := range rn.Peers {
        rn.NextIndex[i] = len(rn.Log)
        rn.MatchIndex[i] = 0
    }

    // Send initial heartbeats
    rn.broadcastHeartbeats()
}
