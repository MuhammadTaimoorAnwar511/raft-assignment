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

// sendHeartbeat sends AppendEntries with no entries
func (rn *RaftNode) sendHeartbeat() {
	for _, peer := range rn.Peers {
		args := rn.prepareAppendEntriesArgs(nil)
		go func(p string, a internal.AppendEntriesArgs) {
			reply := rn.sendAppendEntriesRPC(p, a)

			rn.mu.Lock()
			defer rn.mu.Unlock()

			if reply.Term > rn.currentTerm {
				rn.becomeFollower(reply.Term)
			}
		}(peer, args)
	}
}

// replicateLog is called when we appended a new command locally
func (rn *RaftNode) replicateLog(entries []internal.LogEntry) {
	// We already appended to local log in Propose()
	// Now let's push them to each follower
	for _, p := range rn.Peers {
		go rn.replicateToPeer(p)
	}
}

func (rn *RaftNode) replicateToPeer(peer string) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	for rn.state == internal.Leader {
		nextIdx := rn.nextIndex[peer]
		if nextIdx <= 0 {
			nextIdx = 1
		}

		if nextIdx > rn.LastLogIndex() {
			// follower is up-to-date
			break
		}

		sendEntries := rn.log[nextIdx-1:]
		args := rn.prepareAppendEntriesArgs(sendEntries)
		args.PrevLogIndex = nextIdx - 1
		if args.PrevLogIndex > 0 {
			args.PrevLogTerm = rn.log[args.PrevLogIndex-1].Term
		}

		reply := rn.sendAppendEntriesRPC(peer, args)
		if reply.Term > rn.currentTerm {
			rn.becomeFollower(reply.Term)
			return
		}
		if reply.Success {
			// update matchIndex + nextIndex
			lastIndex := sendEntries[len(sendEntries)-1].Index
			rn.matchIndex[peer] = lastIndex
			rn.nextIndex[peer] = lastIndex + 1
			rn.updateCommitIndex()
			break
		} else {
			// follower log mismatch => fallback
			rn.nextIndex[peer] = reply.NextIndex
			if rn.nextIndex[peer] < 1 {
				rn.nextIndex[peer] = 1
			}
		}
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

	reply.Term = rn.currentTerm
	if args.Term < rn.currentTerm {
		reply.Success = false
		reply.NextIndex = rn.LastLogIndex() + 1
		return
	}

	if args.Term > rn.currentTerm {
		rn.becomeFollower(args.Term)
	} else if rn.state != internal.Follower {
		rn.state = internal.Follower
	}

	// we got a heartbeat => reset election timer
	rn.electionResetEvent = time.Now()

	// check log consistency
	if args.PrevLogIndex > 0 {
		if args.PrevLogIndex > len(rn.log) {
			reply.Success = false
			reply.NextIndex = rn.LastLogIndex() + 1
			return
		}
		prevTerm := rn.log[args.PrevLogIndex-1].Term
		if prevTerm != args.PrevLogTerm {
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
				rn.log = rn.log[:idx-1]
			}
		}
		if idx-1 >= len(rn.log) {
			rn.log = append(rn.log, entry)
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
		rn.applyEntries()
	}

	reply.Success = true
	reply.NextIndex = rn.LastLogIndex() + 1
	return
}

func (rn *RaftNode) findConflictIndex(conflictTerm int, conflictIndex int) int {
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
			rn.applyEntries()
			break
		}
	}
}

func (rn *RaftNode) applyEntries() {
	for rn.lastApplied < rn.commitIndex {
		rn.lastApplied++
		entry := rn.log[rn.lastApplied-1]
		if rn.applyCallback != nil {
			rn.applyCallback(entry.Command)
		}
	}
}
