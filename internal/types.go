package internal

// You can define any shared structs, interfaces, or constants here.
// For example, you might define message types for RPC requests/responses.

// RequestVoteArgs is a potential struct for your RequestVote RPC arguments
type RequestVoteArgs struct {
    Term         int
    CandidateID  int
    LastLogIndex int
    LastLogTerm  int
}

// RequestVoteReply is a potential struct for your RequestVote response
type RequestVoteReply struct {
    Term        int
    VoteGranted bool
}

// AppendEntriesArgs is a potential struct for AppendEntries RPC
type AppendEntriesArgs struct {
    Term         int
    LeaderID     int
    PrevLogIndex int
    PrevLogTerm  int
    Entries      []interface{} // or a custom type for log entries
    LeaderCommit int
}

// AppendEntriesReply is a potential struct for the AppendEntries response
type AppendEntriesReply struct {
    Term    int
    Success bool
}
