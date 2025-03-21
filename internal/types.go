package internal

// CommandType is an enum-like type for commands that the KV store supports.
type CommandType int

const (
	CommandPut CommandType = iota
	CommandAppend
	CommandGet
)

// Command is the structure of a command replicated via Raft.
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

// RaftState represents the node's current state in the Raft protocol.
type RaftState int

const (
	Follower RaftState = iota
	Candidate
	Leader
)

// RequestVoteArgs ...
type RequestVoteArgs struct {
	Term         int
	CandidateID  string
	LastLogIndex int
	LastLogTerm  int
}

// RequestVoteReply ...
type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

// AppendEntriesArgs ...
type AppendEntriesArgs struct {
	Term         int
	LeaderID     string
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

// AppendEntriesReply ...
type AppendEntriesReply struct {
	Term      int
	Success   bool
	NextIndex int
}
