package raft

import (
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/MuhammadTaimoorAnwar511/raft-assignment/internal"
)

// RaftService provides the methods for net/rpc calls.
type RaftService struct {
	rn *RaftNode
}

// RequestVote RPC method.
func (s *RaftService) RequestVote(args *internal.RequestVoteArgs, reply *internal.RequestVoteReply) error {
	*reply = s.rn.handleRequestVote(*args)
	return nil
}

// AppendEntries RPC method.
func (s *RaftService) AppendEntries(args *internal.AppendEntriesArgs, reply *internal.AppendEntriesReply) error {
	*reply = s.rn.handleAppendEntries(*args)
	return nil
}

// startRPCServer registers RaftService and begins listening for RPC calls.
func (rn *RaftNode) startRPCServer() error {
	rpcSrv := &RaftService{rn: rn}
	if err := rpc.Register(rpcSrv); err != nil {
		return err
	}

	l, err := net.Listen("tcp", rn.Address)
	if err != nil {
		return err
	}
	rn.listener = l

	log.Printf("[Node %s] RPC server listening on %s\n", rn.ID, rn.Address)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-rn.stopCh:
					return // we're shutting down
				default:
					log.Printf("[Node %s] accept error: %v", rn.ID, err)
					continue
				}
			}
			go rpc.ServeConn(conn)
		}
	}()
	return nil
}

// handleRequestVote wraps the logic for receiving a RequestVote RPC (like election.go's logic).
func (rn *RaftNode) handleRequestVote(args internal.RequestVoteArgs) internal.RequestVoteReply {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	reply := internal.RequestVoteReply{
		Term:        rn.currentTerm,
		VoteGranted: false,
	}
	// 1) If caller's term < ours, reject
	if args.Term < rn.currentTerm {
		return reply
	}
	// 2) If caller's term > ours, become follower
	if args.Term > rn.currentTerm {
		rn.becomeFollower(args.Term)
	}

	// If we haven't voted yet or we voted for this candidate
	if (rn.votedFor == "" || rn.votedFor == args.CandidateID) &&
		args.LastLogTerm >= rn.lastLogTerm() && args.LastLogIndex >= rn.lastLogIndex() {
		reply.VoteGranted = true
		rn.votedFor = args.CandidateID
		rn.electionResetEvent = nowForElection() // We "heard" from a candidate
	}
	reply.Term = rn.currentTerm
	return reply
}

// sendRequestVoteRPC calls RequestVote on the peer.
func (rn *RaftNode) sendRequestVoteRPC(peerAddr string, args internal.RequestVoteArgs) internal.RequestVoteReply {
	reply := internal.RequestVoteReply{}
	client, err := rpc.Dial("tcp", peerAddr)
	if err != nil {
		// Could not connect => treat as no vote
		return reply
	}
	defer client.Close()

	callErr := client.Call("RaftService.RequestVote", &args, &reply)
	if callErr != nil {
		// network or RPC error => treat as no vote
		return reply
	}
	return reply
}

// sendAppendEntriesRPC calls AppendEntries on the peer.
func (rn *RaftNode) sendAppendEntriesRPC(peerAddr string, args internal.AppendEntriesArgs) internal.AppendEntriesReply {
	reply := internal.AppendEntriesReply{}
	client, err := rpc.Dial("tcp", peerAddr)
	if err != nil {
		return reply
	}
	defer client.Close()

	callErr := client.Call("RaftService.AppendEntries", &args, &reply)
	if callErr != nil {
		return reply
	}
	return reply
}

// Just for consistent time usage
func nowForElection() time.Time {
	return time.Now()
}
