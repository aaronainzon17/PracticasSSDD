package datastructures

type LogEntry struct {
	Command interface{}
	Term    int
}

//Estructuras del AppendEntries
type ArgsAppendEntries struct {
	Term     int
	LeaderId int

	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type RespuestaAppendEntries struct {
	Term    int
	Success bool
}

//Estructuras del RequestVote
type ArgsPeticionVoto struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type RespuestaPeticionVoto struct {
	Term        int
	VoteGranted bool
}
