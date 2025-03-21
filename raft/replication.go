package raft

import "fmt"

// broadcastHeartbeats sends out empty AppendEntries to maintain leadership
func (rn *RaftNode) broadcastHeartbeats() {
    if rn.State != Leader {
        return
    }
    fmt.Printf("[Node %d] Broadcasting heartbeats...\n", rn.ID)
    for i, peer := range rn.Peers {
        if i == rn.ID {
            continue // don't send to self
        }
        go rn.sendHeartbeat(peer)
    }
}

// sendHeartbeat simulates sending an AppendEntries RPC with no new log entries
func (rn *RaftNode) sendHeartbeat(peer string) {
    rn.mu.Lock()
    defer rn.mu.Unlock()

    // Construct the AppendEntries call
    // Typically includes leaderTerm, leaderId, prevLogIndex, prevLogTerm, entries, leaderCommit

    // For now just a debug print
    fmt.Printf("[Node %d] sending heartbeat to %s\n", rn.ID, peer)
}

// In a real scenario, you'd also implement a function like sendAppendEntries(...) for actual log replication.
