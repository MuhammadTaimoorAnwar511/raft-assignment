package raft

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/MuhammadTaimoorAnwar511/raft-assignment/internal"
)

func (rn *RaftNode) runElectionTimer() {
	defer rn.wg.Done()

	for {
		timeout := rn.randomElectionTimeout()

		time.Sleep(timeout)

		rn.mu.Lock()
		if rn.state == internal.Leader {
			rn.mu.Unlock()
			continue
		}

		if time.Since(rn.electionResetEvent) >= timeout {
			fmt.Printf("[Node %s] No leader detected, initiating election\n", rn.ID)
			rn.startElection()
		}
		rn.mu.Unlock()

		select {
		case <-rn.stopCh:
			fmt.Printf("[Node %s] Election timer stopped.\n", rn.ID)
			return
		default:
		}
	}
}

func (rn *RaftNode) startElection() {
	rn.currentTerm++
	rn.state = internal.Candidate
	rn.votedFor = rn.ID
	termStarted := rn.currentTerm

	fmt.Printf("[Node %s] Starting election for term %d\n", rn.ID, termStarted)

	votesReceived := int32(1) // we vote for ourselves

	for _, peerAddr := range rn.Peers {
		go func(addr string) {
			args := internal.RequestVoteArgs{
				Term:         termStarted,
				CandidateID:  rn.ID,
				LastLogIndex: rn.LastLogIndex(),
				LastLogTerm:  rn.LastLogTerm(),
			}
			reply := rn.sendRequestVoteRPC(addr, args)

			rn.mu.Lock()
			defer rn.mu.Unlock()

			if rn.state != internal.Candidate || rn.currentTerm != termStarted {
				return
			}
			if reply.Term > rn.currentTerm {
				fmt.Printf("[Node %s] Higher term detected (%d), reverting to follower\n", rn.ID, reply.Term)
				rn.becomeFollower(reply.Term)
				return
			}
			if reply.VoteGranted {
				fmt.Printf("[Node %s] Received vote from %s\n", rn.ID, addr)
				atomic.AddInt32(&votesReceived, 1)
				if int(atomic.LoadInt32(&votesReceived)) > len(rn.Peers)/2 {
					fmt.Printf("[Node %s] Majority achieved with %d votes\n", rn.ID, votesReceived)
					rn.becomeLeader()
				}
			}
		}(peerAddr)
	}
}

func (rn *RaftNode) becomeFollower(newTerm int) {
	rn.currentTerm = newTerm
	rn.state = internal.Follower
	rn.votedFor = ""
	fmt.Printf("[Node %s] Became Follower in term %d\n", rn.ID, newTerm)
}

func (rn *RaftNode) becomeLeader() {
	rn.state = internal.Leader
	fmt.Printf("[Node %s] Became Leader in term %d\n", rn.ID, rn.currentTerm)

	for _, p := range rn.Peers {
		rn.nextIndex[p] = rn.LastLogIndex() + 1
		rn.matchIndex[p] = 0
	}

	// Send an immediate heartbeat
	rn.sendHeartbeat()
	fmt.Printf("[Node %s] Sent initial heartbeat to all peers\n", rn.ID)
}

func (rn *RaftNode) randomElectionTimeout() time.Duration {
	// typical 150-300ms range
	return time.Duration(150+randIntn(150)) * time.Millisecond
}

func randIntn(n int) int {
	return int(time.Now().UnixNano() % int64(n))
}
