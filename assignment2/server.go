package main

//return values
//get output of voterequest for candidate
type VoteReqEv struct {
	candidateId int
	term int
	lastLogIndex int
	lastLogTerm int
	// etc
}

type Alarm struct{
	t int
}

type Timeout struct {
}

type AppendEntriesReqEv struct {
	term int
	prevLogIndex int
	prevLogTerm int
	entry string
	leaderid int
	leaderCommit int //leader's commited index
	// etc
}

type AppendEntriesResEv struct {
	fromid int
	term int
	status bool
}

type VoteResEv struct{
	fromid int
	term int
	status bool
}

type StateMachine struct {
	state string
	id int // server id
	peers []int // other server ids
	term int //term number
	logs [1024]string //log entries
	terms [1024]int //contains termnumbers for log entries
	lognumber [1024]int //contains logentrynumber
	prevLogTerm int
	index int //array index
	leaderid int //currentleaderid
	commitIndex int //highest log entry commited
	//leaderCommit int //leader's commit index
	votedFor int //candidate recieving vote entry
	lastApplied int //last log entry
	nextIndex int //for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
	matchIndex int //for each server, index of highest log entry known to be replicated on server
	// etc
}

func main() {
	// testing testing
	//var sm1 StateMachine
	sm := StateMachine{state: "f",id:1,term: 1,prevLogTerm: 0, index: 0,leaderid: 2, commitIndex: 0, votedFor:-1 ,lastApplied: 0  }
	sm.ProcessEvent(AppendEntriesReqEv{term : 1, prevLogIndex: 0, prevLogTerm: 0, entry: "add 2 5", leaderid: 2, leaderCommit : 0})
	sm.ProcessEvent(VoteReqEv{term: 2,candidateId:2 , lastLogTerm:1, lastLogIndex:1 })
	sm.ProcessEvent(Timeout{})
} 

func (sm *StateMachine) ProcessEvent (ev interface{}) (interface{},interface{}){
	switch ev.(type) {
	case AppendEntriesReqEv:
		if sm.state=="c" || sm.state=="l"{
			sm.state = "f"
		}
		cmd := ev.(AppendEntriesReqEv)
		if sm.term < cmd.term{
			return AppendEntriesResEv{fromid: sm.id, term: sm.term, status:false},Alarm{}
		}
		if sm.index != cmd.prevLogIndex || sm.prevLogTerm != cmd.prevLogTerm {
			return AppendEntriesResEv{fromid: sm.id, term: sm.term, status:false},Alarm{}
		}

		for sm.index>0 && sm.lognumber[sm.index-1] >= cmd.prevLogIndex+1 && sm.term < cmd.prevLogTerm+1{
			sm.index= sm.index-1
		}
		sm.logs[sm.index] = cmd.entry
		sm.lognumber[sm.index] = cmd.prevLogIndex+1
		sm.terms[sm.index] = cmd.prevLogTerm+1
		sm.prevLogTerm = sm.term
		sm.lastApplied = sm.index+1
		sm.term = cmd.term
		sm.index = sm.index+1

		if cmd.leaderCommit > sm.commitIndex{
			if cmd.leaderCommit > sm.index-1{
				sm.commitIndex = sm.index-1
			}else{
				sm.commitIndex = cmd.leaderCommit
			}
		}
		return AppendEntriesResEv{fromid: sm.id, term: sm.term, status:true},Alarm{}

	case VoteReqEv:
		cmd := ev.(VoteReqEv)
		//var response
		if(sm.state!="l" || sm.state !="c"){
			if cmd.term < sm.term{
				return VoteResEv{fromid: sm.id, term:sm.term, status:false},Alarm{}
			} 
			if cmd.lastLogIndex < sm.lognumber[sm.index-1]{
				return VoteResEv{fromid: sm.id, term:sm.term, status:false},Alarm{}
			}
			if sm.votedFor == -1 || sm.votedFor == cmd.candidateId{
				return VoteResEv{fromid: sm.id, term:sm.term, status:true},Alarm{}
			}

		}
	case Timeout:
		if(sm.state=="f"){
			sm.state = "c"
			sm.term = sm.term+1
			sm.votedFor = sm.id
			return VoteReqEv{term: sm.term,candidateId: sm.id, lastLogIndex: sm.lognumber[sm.index-1], lastLogTerm: sm.terms[sm.index-1]},Alarm{}
		}
		if(sm.state=="c"){
			sm.term = sm.term+1
			sm.votedFor = sm.id
			return VoteReqEv{term: sm.term,candidateId: sm.id, lastLogIndex: sm.lognumber[sm.index-1], lastLogTerm: sm.terms[sm.index-1]},Alarm{}
		}
		if(sm.state == "l"){
			if sm.matchIndex < sm.index{
				for sm.matchIndex < sm.index{
					return AppendEntriesReqEv{term : sm.term, prevLogIndex: sm.lastApplied, prevLogTerm: sm.prevLogTerm, entry: sm.logs[sm.index], leaderid: sm.id, leaderCommit : sm.commitIndex},Alarm{}
					sm.matchIndex = sm.matchIndex+1
				}
			}
		}
	// other cases
	default: println ("Unrecognized")

	}
	return Timeout{},Alarm{}
}