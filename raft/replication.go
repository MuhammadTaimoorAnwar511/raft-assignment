package raft

import (
	"fmt"
	"time"

	"github.com/MuhammadTaimoorAnwar511/raft-assignment/internal"
)

func (rn *RaftNode) runHeartbeatLoop() {
	defer rn.wg.Done()

	heartbeatInterval := 50 * time.Millisecond
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	fmt.Printf("[Node %s] Starting heartbeat loop.\n", rn.ID)

	for {
		select {
		case <-ticker.C:
			rn.mu.Lock()
			if rn.state == internal.Leader {
				rn.sendHeartbeat()
			}
			rn.mu.Unlock()
		case <-rn.stopCh:
			fmt.Printf("[Node %s] Heartbeat loop stopped.\n", rn.ID)
			return
		}
	}
}

// sendHeartbeat sends AppendEntries with actual log entries to help followers catch up
func (rn *RaftNode) sendHeartbeat() {
	//fmt.Printf("[Node %s] Sending heartbeat to peers.\n", rn.ID)
	for _, peer := range rn.Peers {
		go rn.replicateToPeer(peer)
	}
}

func (rn *RaftNode) replicateToPeer(peer string) {
	rn.mu.Lock()
	nextIdx := rn.nextIndex[peer]
	if nextIdx <= 0 {
		nextIdx = 1
	}
	var sendEntries []internal.LogEntry
	if nextIdx <= rn.LastLogIndex() {
		sendEntries = rn.log[nextIdx-1:]
	} else {
		sendEntries = nil // Still send heartbeat
	}
	args := rn.prepareAppendEntriesArgs(sendEntries)
	args.PrevLogIndex = nextIdx - 1
	if args.PrevLogIndex > 0 && args.PrevLogIndex <= len(rn.log) {
		args.PrevLogTerm = rn.log[args.PrevLogIndex-1].Term
	}
	rn.mu.Unlock()

	reply := rn.sendAppendEntriesRPC(peer, args)

	rn.mu.Lock()
	defer rn.mu.Unlock()

	if reply.Term > rn.currentTerm {
		fmt.Printf("[Node %s] Received higher term from peer %s, reverting to follower.\n", rn.ID, peer)
		rn.becomeFollower(reply.Term)
		return
	}

	if reply.Success && len(sendEntries) > 0 {
		lastIndex := sendEntries[len(sendEntries)-1].Index
		rn.matchIndex[peer] = lastIndex
		rn.nextIndex[peer] = lastIndex + 1
		fmt.Printf("[Node %s] Successfully replicated to peer %s, updated nextIndex to %d.\n", rn.ID, peer, rn.nextIndex[peer])
		rn.updateCommitIndex()
	} else if !reply.Success {
		// Compute the effective new nextIndex value
		newIndex := reply.NextIndex
		if newIndex < 1 {
			newIndex = 1
		}
		// Only print if the effective value changes
		if rn.nextIndex[peer] != newIndex {
			fmt.Printf("[Node %s] Replication failed for peer %s, setting nextIndex to %d.\n", rn.ID, peer, newIndex)
		}
		rn.nextIndex[peer] = newIndex
	}
}

func (rn *RaftNode) prepareAppendEntriesArgs(entries []internal.LogEntry) internal.AppendEntriesArgs {
	return internal.AppendEntriesArgs{
		Term:         rn.currentTerm,
		LeaderID:     rn.ID,
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries:      entries,
		LeaderCommit: rn.commitIndex,
	}
}

// handleAppendEntries is called via net/rpc from a leader
func (rn *RaftNode) handleAppendEntries(args internal.AppendEntriesArgs) (reply internal.AppendEntriesReply) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// Only print detailed info when actual log entries are being appended (i.e. not for simple heartbeats)
	if len(args.Entries) > 0 {
		fmt.Printf("[Node %s] Handling AppendEntries RPC with %d log entries.\n", rn.ID, len(args.Entries))
	}

	reply.Term = rn.currentTerm
	if args.Term < rn.currentTerm {
		fmt.Printf("[Node %s] AppendEntries RPC term %d is less than current term %d, rejecting request.\n", rn.ID, args.Term, rn.currentTerm)
		reply.Success = false
		reply.NextIndex = rn.LastLogIndex() + 1
		return
	}

	if args.Term > rn.currentTerm {
		fmt.Printf("[Node %s] AppendEntries RPC term %d is higher than current term %d, \033[33mbecoming follower\033[0m.\n", rn.ID, args.Term, rn.currentTerm)
		rn.becomeFollower(args.Term)
	} else if rn.state != internal.Follower {
		fmt.Printf("[Node %s] AppendEntries RPC received, but node state is %d instead of Follower, setting state to Follower.\n", rn.ID, rn.state)
		rn.state = internal.Follower
	}

	// we got a heartbeat => reset election timer
	rn.electionResetEvent = time.Now()

	// check log consistency
	if args.PrevLogIndex > 0 {
		if args.PrevLogIndex > len(rn.log) {
			fmt.Printf("[Node %s] Log inconsistency: PrevLogIndex %d out of range. Log length: %d.\n", rn.ID, args.PrevLogIndex, len(rn.log))
			reply.Success = false
			reply.NextIndex = rn.LastLogIndex() + 1
			return
		}
		prevTerm := rn.log[args.PrevLogIndex-1].Term
		if prevTerm != args.PrevLogTerm {
			fmt.Printf("[Node %s] Log inconsistency at PrevLogIndex %d: expected term %d, got %d. Finding conflict index.\n", rn.ID, args.PrevLogIndex, prevTerm, args.PrevLogTerm)
			reply.Success = false
			reply.NextIndex = rn.findConflictIndex(prevTerm, args.PrevLogIndex)
			return
		}
	}

	// append new entries
	for i, entry := range args.Entries {
		idx := args.PrevLogIndex + i + 1
		if idx-1 < len(rn.log) {
			if rn.log[idx-1].Term != entry.Term {
				fmt.Printf("[Node %s] Conflict at index %d: log term mismatch. Truncating log.\n", rn.ID, idx)
				rn.log = rn.log[:idx-1]
			}
		}
		if idx-1 >= len(rn.log) {
			rn.log = append(rn.log, entry)
			fmt.Printf("[Node %s] Appended log entry at index %d.\n", rn.ID, idx)
		}
	}

	// update commit index
	if args.LeaderCommit > rn.commitIndex {
		lastNewIndex := args.PrevLogIndex + len(args.Entries)
		if args.LeaderCommit < lastNewIndex {
			rn.commitIndex = args.LeaderCommit
		} else {
			rn.commitIndex = lastNewIndex
		}
		fmt.Printf("[Node %s] Updated commit index to %d. Applying entries.\n", rn.ID, rn.commitIndex)
		rn.applyEntries()
	}

	reply.Success = true
	reply.NextIndex = rn.LastLogIndex() + 1
	return
}

func (rn *RaftNode) findConflictIndex(_ int, conflictIndex int) int {
	return conflictIndex - 1 // naive fallback
}

func (rn *RaftNode) updateCommitIndex() {
	for n := rn.LastLogIndex(); n > rn.commitIndex; n-- {
		count := 1 // leader includes itself
		for _, p := range rn.Peers {
			if rn.matchIndex[p] >= n {
				count++
			}
		}
		if count > len(rn.Peers)/2 && rn.log[n-1].Term == rn.currentTerm {
			rn.commitIndex = n
			fmt.Printf("[Node %s] Commit index updated to %d. Applying entries.\n", rn.ID, rn.commitIndex)
			rn.applyEntries()
			break
		}
	}
}

func (rn *RaftNode) applyEntries() {
	for rn.lastApplied < rn.commitIndex {
		rn.lastApplied++
		entry := rn.log[rn.lastApplied-1]
		fmt.Printf("[Node %s] Applying log entry with index %d.\n", rn.ID, entry.Index)
		if rn.applyCallback != nil {
			rn.applyCallback(entry.Command)
		}
	}
}
