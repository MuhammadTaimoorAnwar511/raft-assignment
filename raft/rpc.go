package raft

import (
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/MuhammadTaimoorAnwar511/raft-assignment/internal"
)

type RaftService struct {
	rn *RaftNode
}

// RequestVote is invoked remotely by other nodes
func (s *RaftService) RequestVote(args *internal.RequestVoteArgs, reply *internal.RequestVoteReply) error {
	*reply = s.rn.handleRequestVote(*args)
	return nil
}

// AppendEntries is invoked remotely by other nodes
func (s *RaftService) AppendEntries(args *internal.AppendEntriesArgs, reply *internal.AppendEntriesReply) error {
	*reply = s.rn.handleAppendEntries(*args)
	return nil
}

func (rn *RaftNode) startRPCServer() error {
	service := &RaftService{rn: rn}
	if err := rpc.Register(service); err != nil {
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
					return
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

// handleRequestVote processes an incoming RequestVote RPC
func (rn *RaftNode) handleRequestVote(args internal.RequestVoteArgs) internal.RequestVoteReply {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	reply := internal.RequestVoteReply{
		Term:        rn.currentTerm,
		VoteGranted: false,
	}

	if args.Term < rn.currentTerm {
		// candidate is behind
		return reply
	}
	if args.Term > rn.currentTerm {
		rn.becomeFollower(args.Term)
	}

	// If we haven't voted yet or we voted for this candidate,
	// and candidate's log is at least as up-to-date as ours
	upToDate := (args.LastLogTerm > rn.LastLogTerm()) ||
		(args.LastLogTerm == rn.LastLogTerm() && args.LastLogIndex >= rn.LastLogIndex())

	if (rn.votedFor == "" || rn.votedFor == args.CandidateID) && upToDate {
		reply.VoteGranted = true
		rn.votedFor = args.CandidateID
		rn.electionResetEvent = nowForElection()
	}
	reply.Term = rn.currentTerm
	return reply
}

// sendRequestVoteRPC calls RequestVote on the peer
func (rn *RaftNode) sendRequestVoteRPC(peerAddr string, args internal.RequestVoteArgs) internal.RequestVoteReply {
	reply := internal.RequestVoteReply{}
	client, err := rpc.Dial("tcp", peerAddr)
	if err != nil {
		return reply
	}
	defer client.Close()

	if err := client.Call("RaftService.RequestVote", &args, &reply); err != nil {
		return reply
	}
	return reply
}

// sendAppendEntriesRPC calls AppendEntries on the peer
func (rn *RaftNode) sendAppendEntriesRPC(peerAddr string, args internal.AppendEntriesArgs) internal.AppendEntriesReply {
	reply := internal.AppendEntriesReply{}
	client, err := rpc.Dial("tcp", peerAddr)
	if err != nil {
		return reply
	}
	defer client.Close()

	if err := client.Call("RaftService.AppendEntries", &args, &reply); err != nil {
		return reply
	}
	return reply
}

// Utility to unify 'time.Now()' usage
func nowForElection() time.Time {
	return time.Now()
}
