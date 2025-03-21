package internal

// CommandType is an enum-like type for commands that the KV store supports.
type CommandType int

const (
	CommandPut CommandType = iota
	CommandAppend
	CommandGet // Although "Get" typically doesn't need to go through Raft if it doesn't change state.
)

// Command is the structure of a command replicated via Raft.
// For example, "Put" or "Append" with a Key/Value.
type Command struct {
	Type  CommandType
	Key   string
	Value string
}

// LogEntry is a single entry in the Raft log.
type LogEntry struct {
	Term    int
	Index   int
	Command Command
}

// RaftState represents the node's current state in the Raft protocol (Follower, Candidate, Leader).
type RaftState int

const (
	Follower RaftState = iota
	Candidate
	Leader
)

// RequestVoteArgs corresponds to the "Request Vote" RPC arguments in the Raft paper.
type RequestVoteArgs struct {
	Term         int
	CandidateID  string
	LastLogIndex int
	LastLogTerm  int
}

// RequestVoteReply corresponds to the "Request Vote" RPC reply in the Raft paper.
type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

// AppendEntriesArgs corresponds to the "Append Entries" RPC arguments in the Raft paper.
type AppendEntriesArgs struct {
	Term         int
	LeaderID     string
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

// AppendEntriesReply corresponds to the "Append Entries" RPC reply in the Raft paper.
type AppendEntriesReply struct {
	Term      int
	Success   bool
	NextIndex int
}
