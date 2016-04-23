package main
import(
	"fmt"
	"math"
	"math/rand"
)
//return values
//get output of voterequest for candidate
type VoteReq struct {
	candidateId int
	term int
	lastLogIndex int
	lastLogTerm int
	// etc
}

type Timeout struct {
}

type AppendEntriesReq struct {
	term int
	prevLogIndex int
	prevLogTerm int
	data []string
	leaderid int
	leaderCommit int //leader's commited index
	// etc
}

type AppendEntriesResp struct {
	fromid int
	term int
	lastLogIndex int
	status bool
}

type VoteResp struct{
	fromid int
	term int
	status bool

}


type Append struct{
	data []byte
	clientid int
}

type StateMachine struct {
	state string
	id int // server id
	peers []int // other server ids
	term int //term number
	logs [1024]string //log entries
	terms [1024]int //contains termnumbers for log entries
	lognumber [1024]int //contains log entry number
	prevLogTerm int
	index int //array index
	leaderid int //currentleaderid
	commitIndex int //highest log entry commited
	leaderCommit int //leader's commit index
	votedFor int //candidate recieving vote entry
	lastApplied int //last log entry
	nextIndex map[int]int //for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
	matchIndex [1024]int //for each server, index of highest log entry known to be replicated on server
	// etc
}
type Event interface{

}
type Action interface{

}

type Error struct{
	err string
}

type Send struct{
	peerId int
	event Event
}

type Commit struct{
	index int
	data string
	err Error
}

type Alarm struct{
	t int
}

type LogStore struct{
	index int
	data []byte
} 



func handleAppend(sm *StateMachine,cmd * Append) []Action{
	var action []Action
	var cmt Commit
	if sm.state=="candidate"{
		cmt.data = string(cmd.data[:])
		cmt.err.err = "Wait_For_Election"
		action = append(action,cmt)
	}else if sm.state == "follower"{
		cmt.data = string(cmd.data[:])
		cmt.err.err = "Sent_To_Leader"
		action = append(action,cmt)
		action = append(action, Send{3,Append{[]byte{'s','u','b'},cmd.clientid}})
	}else{
		cmt.data = string(cmd.data[:])
		cmt.index = sm.index+1
		cmt.err.err=""
		action = append(action,cmt)
		action = append(action,LogStore{sm.index+1,cmd.data})
		for i:=0;i<len(sm.peers);i++{
			action = append(action,Send{sm.peers[i],AppendEntriesReq{term:sm.term, prevLogIndex:sm.index,prevLogTerm:sm.term,data:[]string{string(cmd.data)},leaderid:sm.leaderid,leaderCommit:sm.commitIndex}})			
		}
		sm.logs[sm.index] = string(cmd.data[:])
		sm.term = 2
		sm.prevLogTerm = sm.term
		sm.lognumber[sm.index] = sm.index+1
		sm.index = sm.index+1
		for i:=0;i<len(sm.nextIndex);i++{
			sm.nextIndex[sm.peers[i]]=sm.nextIndex[sm.peers[i]]+1
		}
		sm.commitIndex = sm.commitIndex+1
		sm.lastApplied = sm.lastApplied+1
	}
	return action
}

func handleAppendEntriesReq(sm *StateMachine,cmd *AppendEntriesReq) []Action{
	var action []Action
	var timeout int
	timeout = int(rand.Float64()*5+5)
	if (cmd.term<sm.term){
		action = append(action, Send{cmd.leaderid, AppendEntriesResp{sm.id,sm.term,sm.lastApplied,false}})
		return action
	}
	action = append(action,Alarm{timeout})
	if (cmd.prevLogTerm != sm.prevLogTerm || sm.lastApplied !=cmd.prevLogIndex){
		action = append(action, Send{sm.leaderid, AppendEntriesResp{sm.id, sm.term, sm.lastApplied,false}})
	}else{
		for i:=0;i<len(cmd.data);i++{
			action = append(action, LogStore{sm.index,[]byte(sm.logs[i])})
			sm.logs[sm.lastApplied+i] = cmd.data[i]
			sm.terms[sm.lastApplied+i] = cmd.term
			sm.lognumber[sm.lastApplied+i] = sm.index
			sm.index++
		}
		sm.lastApplied = sm.lastApplied+len(cmd.data)
		//fmt.Println(len(cmd.data))
		sm.prevLogTerm = cmd.term
		action = append(action, Send{sm.leaderid, AppendEntriesResp{sm.id, sm.term, sm.lastApplied, false}})
		sm.leaderCommit = cmd.leaderCommit
		sm.commitIndex = int(math.Min(float64(cmd.leaderCommit),float64(sm.lastApplied)))
		var cmt Commit
		cmt.err.err = ""
		fmt.Println(cmd.data[len(cmd.data)-1])
		action = append(action, Commit{sm.commitIndex,cmd.data[len(cmd.data)-1],cmt.err})

	}

	return action
}

func handleAppendEntriesResp(sm *StateMachine,cmd *AppendEntriesReq) []Action{
	var action []Action
	if(sm.state == "leader"){
		
	}
	return action
}

func handleTimeout(sm * StateMachine) []Action{
	var action []Action
	timeout := int(rand.Float64()*5+5)
	action = append(action,Alarm{timeout})
	if(sm.state == "leader"){
		for i:=0;i<len(sm.peers);i++{
			action = append(action,Send{sm.peers[i],AppendEntriesReq{sm.term,sm.nextIndex[sm.peers[i]],sm.terms[sm.nextIndex[sm.peers[i]]],[]string{},sm.leaderid,sm.commitIndex}})
		}
	}else if sm.state=="follower" {
		sm.state = "candidate"
		sm.term++
		sm.votedFor = sm.id
		for i:=0;i<len(sm.peers);i++{
			action = append(action,Send{sm.peers[i],VoteReq{sm.id,sm.term,sm.lastApplied,sm.prevLogTerm}})
		}	
	}
	return action
}

func handleVoteReq(sm * StateMachine,cmd *VoteReq) []Action{
	var action []Action
	if sm.state == "follower"{
			if(cmd.term < sm.term){
				action = append(action,Send{cmd.candidateId,VoteResp{sm.id, sm.term, false}})
				return action
			}
			if(sm.prevLogTerm > cmd.lastLogTerm) || ((sm.prevLogTerm == cmd.lastLogTerm) && (sm.lastApplied > cmd.lastLogIndex)){
				action = append(action, Send{cmd.candidateId,VoteResp{sm.id, sm.term, false}})
				return action
			}
			if(sm.votedFor != 0){
				if(sm.votedFor != cmd.candidateId){
					action = append(action,Send{cmd.candidateId,VoteResp{sm.id, sm.term, false}})
					return action
				}else{
					sm.votedFor = cmd.candidateId
					action = append(action,Send{cmd.candidateId,VoteResp{sm.id, sm.term, true}})
					return action
				}
			}else{
				action = append(action,Send{cmd.candidateId,VoteResp{sm.id, sm.term, true}})
				return action
			}
		}else{
			if(sm.term < cmd.term){
				sm.state = "follower"
				sm.term = cmd.term
				sm.votedFor = 0
				if (cmd.lastLogTerm < sm.prevLogTerm) || ((cmd.lastLogTerm == sm.prevLogTerm) && (sm.lastApplied > cmd.lastLogIndex)){
					action = append(action, Send{cmd.candidateId,VoteResp{sm.id, sm.term, false}})
					return action
				}
				sm.votedFor = cmd.candidateId
				action = append(action, Send{cmd.candidateId,VoteResp{sm.id, sm.term, true}})
				return action
			}
			action = append(action, Send{cmd.candidateId,VoteResp{sm.id, sm.term, false}})
		}
	return action
}

func handleVoteResp(sm * StateMachine,cmd *VoteResp) []Action{
	var action []Action
	return action
}

func (sm *StateMachine) ProcessEvent (ev Event) []Action{
	var action []Action
	switch ev.(type) {
		case Append:
			cmd := ev.(Append)
			action = handleAppend(sm,&cmd)
		case AppendEntriesReq:
			cmd := ev.(AppendEntriesReq)
			action = handleAppendEntriesReq(sm,&cmd)
		case Timeout:
			_ = ev.(Timeout)
			action = handleTimeout(sm)
		case AppendEntriesResp:
			cmd := ev.(AppendEntriesResp)
			action := handleAppendEntriesResp(sm,&cmd)
		case VoteReq:
			cmd := ev.(VoteReq)
			action = handleVoteReq(sm,&cmd)
		case VoteResp:
			cmd := ev.(VoteResp)
			action := handleVoteResp(sm,&cmd)
		default:
			return action
		}
	return action
}



/*func (sm *StateMachine) ProcessEvent (ev interface{}) []Action{
	var action []Action
	switch ev.(type) {
	case AppendEntriesReqEv:
		if sm.state=="c" || sm.state=="l"{
			sm.state = "f"
		}
		cmd := ev.(AppendEntriesReqEv)
		if sm.term < cmd.term{
			action = append(action,AppendEntriesResEv{fromid: sm.id, term: sm.term, status:false})
			action = append(action,Alarm{})
			return action
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
}*/