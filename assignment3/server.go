package main
import(
	"fmt"
	"math"
	"time"
	"math/rand"
	"github.com/cs733-iitb/log"
	"github.com/cs733-iitb/cluster"
	"github.com/cs733-iitb/cluster/mock"
)

type Node struct {
	id int
	leaderid int
	LogDir string
	sm *StateMachine
	timeoutCh *time.Timer
	server *mock.MockServer
	CommitCh chan *Commit
}



type ConfigRaft struct {
	cluster []*NetConfig
	Id int
	LogDir string
	ElectionTimeout  int
	HeartbeatTimeout int
}

type NetConfig struct {
	Id   int
	Host string
	Port int
}
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
	Votes []int
	Log	*log.Log
	lastApplied int //last log entry
	nextIndex map[int]int //for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
	matchIndex map[int]int //for each server, index of highest log entry known to be replicated on server
	HeartbeatTimeout int
	ElectionTimeout int
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

func (rn *Node) CommitChannel() chan *Commit {
	return rn.CommitCh
}

var leaderId int

func respondAppend(sm *StateMachine,temp * Append) []Action{
	var action []Action
	var cmt Commit
	if sm.state=="candidate"{
		cmt.data = string(temp.data[:])
		cmt.err.err = "Wait_For_Election"
		action = append(action,cmt)
	}else if sm.state == "follower"{
		cmt.data = string(temp.data[:])
		cmt.err.err = "Sent_To_Leader"
		action = append(action,cmt)
		action = append(action, Send{3,Append{[]byte{'s','u','b'},temp.clientid}})
	}else{
		cmt.data = string(temp.data[:])
		cmt.index = sm.index+1
		cmt.err.err=""
		action = append(action,cmt)
		action = append(action,LogStore{sm.index+1,temp.data})
		for i:=0;i<len(sm.peers);i++{
			action = append(action,Send{sm.peers[i],AppendEntriesReq{term:sm.term, prevLogIndex:sm.index,prevLogTerm:sm.term,data:[]string{string(temp.data)},leaderid:sm.leaderid,leaderCommit:sm.commitIndex}})			
		}
		sm.logs[sm.index] = string(temp.data[:])
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

func respondAppendEntriesReq(sm *StateMachine,temp *AppendEntriesReq) []Action{
	var action []Action
	var timeout int
	timeout = int(rand.Float64()*5+5)
	if(temp.term<sm.term){
		fmt.Println("fvfbg")
		action = append(action, Send{temp.leaderid, AppendEntriesResp{sm.id,sm.term,sm.lastApplied,false}})
		return action
	}
	action = append(action,Alarm{timeout})
	if(temp.prevLogTerm != sm.prevLogTerm || sm.lastApplied !=temp.prevLogIndex){
		//fmt.Println("fvfbg")
		action = append(action, Send{sm.leaderid, AppendEntriesResp{sm.id, sm.term, sm.lastApplied,false}})
	}else{
		//fmt.Println("fvfbg")
		for i:=0;i<len(temp.data);i++{
			action = append(action, LogStore{sm.index,[]byte(sm.logs[i])})
			sm.logs[sm.lastApplied+i] = temp.data[i]
			sm.terms[sm.lastApplied+i] = temp.term
			sm.lognumber[sm.lastApplied+i] = sm.index
			sm.index++
		}
		sm.lastApplied = sm.lastApplied+len(temp.data)
		//fmt.Println(len(temp.data))
		sm.prevLogTerm = temp.term
		action = append(action, Send{sm.leaderid, AppendEntriesResp{sm.id, sm.term, sm.lastApplied, false}})
		sm.leaderCommit = temp.leaderCommit
		sm.commitIndex = int(math.Min(float64(temp.leaderCommit),float64(sm.lastApplied)))
		var cmt Commit
		cmt.err.err = ""
		fmt.Println(temp.data[len(temp.data)-1])
		action = append(action, Commit{sm.commitIndex,temp.data[len(temp.data)-1],cmt.err})

	}

	return action
}

func respondAppendEntriesResp(sm *StateMachine,temp *AppendEntriesResp) []Action{
	var action []Action
	if(sm.state == "leader"){
		if sm.term < temp.term{
			sm.state = "follower"
			sm.term = temp.term

		}
	if temp.status==false{
			sm.nextIndex[temp.fromid] = temp.lastLogIndex+1
			action = append(action,Send{temp.fromid,AppendEntriesReq{sm.term,sm.lognumber[sm.nextIndex[temp.fromid]-1],sm.terms[sm.nextIndex[temp.fromid]-1],sm.logs[sm.nextIndex[temp.fromid]:sm.index-1],sm.id,sm.leaderCommit}})
			return action
		}else{
			sm.nextIndex[temp.fromid] = temp.lastLogIndex+1
			sm.matchIndex[temp.fromid] = temp.lastLogIndex

		   	if sm.nextIndex[temp.fromid] <sm.index {
				action = append(action,AppendEntriesReq{sm.term,sm.lognumber[sm.nextIndex[temp.fromid]-1],sm.terms[sm.nextIndex[temp.fromid]-1],sm.logs[sm.nextIndex[temp.fromid]:sm.index-1],sm.id,sm.leaderCommit})
			}

			for i:=sm.matchIndex[temp.fromid];i <sm.index;i++{
				count := 1
				for _,v := range sm.matchIndex{
					if v>i{
						count++
					}
				}
				if count>2 && i>sm.commitIndex{
					sm.commitIndex++
					action = append(action,Commit{index:sm.commitIndex,data:sm.logs[sm.commitIndex],})
				}
			}
		}
	}
	return action
}

func respondTimeout(sm * StateMachine) []Action{
	var action []Action
	if(sm.state == "leader"){
		timeout := int(sm.HeartbeatTimeout)
		action = append(action,Alarm{timeout})
		for i:=0;i<len(sm.peers);i++{
			if(sm.peers[i]!=0){
				action = append(action,Send{sm.peers[i],AppendEntriesReq{sm.term,sm.nextIndex[sm.peers[i]],sm.terms[sm.nextIndex[sm.peers[i]]],[]string{},sm.leaderid,sm.commitIndex}})
			}
		}
	}else {
		timeout := int(rand.Float64()*float64(sm.ElectionTimeout))+sm.ElectionTimeout
		action = append(action,Alarm{timeout})
		sm.state = "candidate"
		sm.term++
		sm.votedFor = sm.id
		for i:=0;i<5;i++{
			sm.Votes[i] = 0
		}
		for i:=0;i<len(sm.peers);i++{
			if(sm.peers[i]!=0){
				action = append(action,Send{sm.peers[i],VoteReq{sm.id,sm.term,sm.lastApplied,sm.prevLogTerm}})
			}
		}	
	}
	return action
}

func respondVoteReq(sm * StateMachine,temp *VoteReq) []Action{
	var action []Action
	if sm.state == "follower"{
			if(temp.term < sm.term){
				action = append(action,Send{temp.candidateId,VoteResp{sm.id, sm.term, false}})
				return action
			}
			if(sm.prevLogTerm > temp.lastLogTerm) || ((sm.prevLogTerm == temp.lastLogTerm) && (sm.lastApplied > temp.lastLogIndex)){
				action = append(action, Send{temp.candidateId,VoteResp{sm.id, sm.term, false}})
				return action
			}
			if(sm.votedFor != 0){
				if(sm.votedFor != temp.candidateId){
					action = append(action,Send{temp.candidateId,VoteResp{sm.id, sm.term, false}})
					return action
				}else{
					sm.votedFor = temp.candidateId
					action = append(action,Send{temp.candidateId,VoteResp{sm.id, sm.term, true}})
					return action
				}
			}else{
				action = append(action,Send{temp.candidateId,VoteResp{sm.id, sm.term, true}})
				return action
			}
		}else{
			if(sm.term < temp.term){
				sm.state = "follower"
				sm.term = temp.term
				sm.votedFor = 0
				if (temp.lastLogTerm < sm.prevLogTerm) || ((temp.lastLogTerm == sm.prevLogTerm) && (sm.lastApplied > temp.lastLogIndex)){
					action = append(action, Send{temp.candidateId,VoteResp{sm.id, sm.term, false}})
					return action
				}
				sm.votedFor = temp.candidateId
				action = append(action, Send{temp.candidateId,VoteResp{sm.id, sm.term, true}})
				return action
			}
			action = append(action, Send{temp.candidateId,VoteResp{sm.id, sm.term, false}})
		}
	return action
}

func respondVoteResp(sm * StateMachine,temp *VoteResp) []Action{
	var action []Action
	var timeout int
	action = append(action, Alarm{timeout})
	timeout = int(rand.Float64()*5+5)
	if sm.term<temp.term{
		sm.state = "follower"
		sm.term = temp.term
		
	}
	sm.Votes = make([]int,5)
	if sm.state == "candidate"{
		for i:=0;i<5;i++{
			sm.Votes[sm.peers[i]-1]=0
		}
		sm.Votes[sm.id-1] = 1
		if temp.status == true{	
			sm.Votes[temp.fromid-1] = 1
		}
		if temp.status == true{
			count := 0
			for i:=0;i<5;i++{
				if sm.Votes[i]==1{
					count++
				}
			}
			if count>2{
				sm.state = "leader"
				sm.leaderid = sm.id
				for i:=0;i<len(sm.peers);i++{
					if sm.peers[i]!=sm.id{
					action = append(action, Send{sm.peers[i],AppendEntriesReq{sm.term,sm.lastApplied,sm.lognumber[sm.lastApplied],[]string{},sm.id,sm.commitIndex}})
					}
				} 
			}

		} 
	}
	return action
}


func (sm *StateMachine) ProcessEvent (ev Event) []Action{
	var action []Action
	switch ev.(type) {
		case Append:
			temp := ev.(Append)
			action = respondAppend(sm,&temp)
		case AppendEntriesReq:
			temp := ev.(AppendEntriesReq)
			action = respondAppendEntriesReq(sm,&temp)
		case Timeout:
			_ = ev.(Timeout)
			action = respondTimeout(sm)
		case AppendEntriesResp:
			temp := ev.(AppendEntriesResp)
			action := respondAppendEntriesResp(sm,&temp)
		case VoteReq:
			temp := ev.(VoteReq)
			action = respondVoteReq(sm,&temp)
		case VoteResp:
			temp := ev.(VoteResp)
			action := respondVoteResp(sm,&temp)
		default:
			//return action
		}
	return action
}


func (rn *Node) listenonchannel() {
	s := rn.sm
	for {
		select {
		case e := <-rn.server.Inbox():
			action := s.ProcessEvent(e.Msg)
			if s.leaderid!=-1 && s.state=="leader" {
				rn.leaderid = s.leaderid
			}
			rn.writechannel(action)
		case <-rn.timeoutCh.C:

			action := s.ProcessEvent(Timeout{})
			if s.leaderid!=-1 && s.state== "leader"{
				rn.leaderid = s.leaderid
				fmt.Println(rn.id)
			}
			
			rn.writechannel(action)
		}
	}
}

func (rn *Node) writechannel(ev []Action) {
	for i:=0;i<len(ev);i++{
		switch ev[i].(type){
				case Send:
					temp := ev[i].(Send)
					switch temp.event.(type){
					case AppendEntriesReq:
						Msg := temp.event.(AppendEntriesReq)
						rn.server.Outbox() <-  &cluster.Envelope{Pid:temp.peerId, Msg:Msg}
					case AppendEntriesResp:
						Msg := temp.event.(AppendEntriesResp)
						rn.server.Outbox() <-  &cluster.Envelope{Pid:temp.peerId, Msg:Msg}
					case VoteReq:
						Msg := temp.event.(VoteReq)
						rn.server.Outbox() <-  &cluster.Envelope{Pid:temp.peerId, Msg:Msg}
					case VoteResp:
						Msg := temp.event.(VoteResp)
						rn.server.Outbox() <-  &cluster.Envelope{Pid:temp.peerId, Msg:Msg}
					}
				case Alarm:
					temp := ev[i].(Alarm)
					rn.timeoutCh.Reset(time.Duration(temp.t)*time.Millisecond)
				case Commit:
					temp := ev[i].(Commit)
					rn.CommitChannel() <- &Commit{index:temp.index,data:temp.data}
				default:
					

			}
	}
}