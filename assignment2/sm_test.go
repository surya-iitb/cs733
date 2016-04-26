package main

import (
	"fmt"
	"testing"
)


func TestAppend(t *testing.T){
	sm := StateMachine{state: "leader", id:7, peers:[]int{1,2,3,4,5}, term:2, prevLogTerm:1, index:2, leaderid:7, commitIndex:3, lastApplied:3, nextIndex: map[int]int{1:4,2:4,3:4,4:4,5:4} }
	sm.logs[0]="jmp"
	sm.logs[1]="jmpz"
	sm.logs[2]="add"
	for i:=0;i<3;i++{
		sm.terms[i]=1
		sm.lognumber[i]=i+1
	}
	expectedIndex := sm.index
	expectedData :="sub"
	numAppendEntries := 0
	//check append for leader
	action := sm.ProcessEvent(Append{[]byte{'s','u','b'},10})

	for i :=0;i<len(action);i++{
		switch action[i].(type){
		case Commit:
			temp := action[i].(Commit)
			//expectedData = temp.data
			//err = temp.err
			//expectedIndex = temp.index
			expect(t, string(temp.index),string(expectedIndex+1))
			expect(t, temp.err.err,string(""))
			expect(t, expectedData,temp.data)
		case LogStore:
			temp := action[i].(LogStore)
			expect(t, string(temp.index),string(expectedIndex+1))
			//expect(t, temp.err.err,string(""))
			expect(t, expectedData,string(temp.data))
		case Send:
			numAppendEntries++
		}
	}
	//fmt.Print(numAppendEntries)
	expect(t, string(numAppendEntries),string(5))

	sm.state = "follower"
	sm.term = 3
	sm.prevLogTerm = 2
	sm.leaderid = 5
	action = sm.ProcessEvent(Append{[]byte{'s','u','b'},11})
	for i :=0;i<len(action);i++{
		switch action[i].(type){
		case Commit:
			temp := action[i].(Commit)
			expect(t, temp.err.err,string("Sent_To_Leader"))
			expect(t, expectedData,temp.data)
		case Send:
			temp := action[i].(Send)
			expect(t,string(temp.peerId),string(3))
			z := temp.event.(Append)
			expect(t,string(z.data),expectedData)
		}
	}
	sm.state = "candidate"
	sm.term = 4
	sm.prevLogTerm = 3
	action = sm.ProcessEvent(Append{[]byte{'s','u','b'},12})
	temp := action[0].(Commit)
	expect(t, temp.err.err,string("Wait_For_Election"))
	expect(t, expectedData,temp.data)
	
}

func TestAppendEntriesReq(t *testing.T){
	sm := StateMachine{state: "follower", id:1, peers:[]int{1,2,3,4,5}, term:3, prevLogTerm:2, index:3, leaderid:7, commitIndex:2, leaderCommit:2, lastApplied:2, nextIndex: map[int]int{1:3,2:3,3:3,4:3,5:3} }
	sm.logs[0]="jmp"
	sm.logs[1]="jmpz"
	sm.terms[0]=2
	sm.terms[1]=2
	sm.lognumber[0]=1
	sm.lognumber[1]=2
	alarm := 0
	numlogstore := 0
	var logs []string
	logs = []string{"add","sub"}
	var action []Action
	action = sm.ProcessEvent(AppendEntriesReq{3,sm.lastApplied,2,logs,7,4})
	for i :=0;i<len(action);i++{
		switch action[i].(type){
			case Send:
				temp := action[i].(Send)
				expect(t,string(temp.peerId),string(sm.leaderid))
				z := temp.event.(AppendEntriesResp)
				expect(t,string(z.term),string(sm.term))
				expect(t,string(z.fromid),string(1))
				//fmt.Print(z.lastLogIndex)
				expect(t,string(z.lastLogIndex),string(4))
			case Alarm:
				alarm++
			case LogStore:
				numlogstore++
			case Commit:
				temp := action[i].(Commit)
				expect(t,string(temp.index),string(4))
				expect(t,temp.data,"sub")

		}	
	}


	expect(t,string(alarm),string(1))
	expect(t,string(numlogstore),string(2))
}

func TestAppendEntriesResp(t *testing.T){
	sm := StateMachine{state: "leader", id:1, peers:[]int{1,2,3,4,5}, term:3, prevLogTerm:2, index:3, leaderid:1, commitIndex:2, leaderCommit:2, lastApplied:2, nextIndex: map[int]int{1:3,2:3,3:3,4:3,5:3}, matchIndex: map[int]int{1:2,2:2,3:2,4:2,5:2}}
	sm.logs[0]="jmp"
	sm.logs[1]="jmpz"
	sm.terms[0]=2
	sm.terms[1]=2
	sm.lognumber[0]=1
	sm.lognumber[1]=2
	action := sm.ProcessEvent(AppendEntriesResp{2,3,1,true})
	//fmt.Println(len(action))
	expect(t,string(len(action)),string(1))	

}

func TestTimeout(t *testing.T){
	sm := StateMachine{state: "follower", id:1, peers:[]int{1,2,3,4,7}, term:3, prevLogTerm:2, index:3, leaderid:7, commitIndex:2, leaderCommit:2, lastApplied:2, nextIndex: map[int]int{1:3,2:3,3:3,4:3,5:3} }
	sm.logs[0]="jmp"
	sm.logs[1]="jmpz"
	sm.terms[0]=2
	sm.terms[1]=2
	sm.lognumber[0]=1
	sm.lognumber[1]=2
	var action []Action
	alarm := 0
	numevents := 0
	action = sm.ProcessEvent(Timeout{})
	for i :=0;i<len(action);i++{
		switch action[i].(type){
			case Alarm:
				alarm++
			case Send:
				numevents++
		}	
	}
	//fmt.Print(len(action))
	expect(t,string(alarm),string(1))
	expect(t,string(numevents),string(5))
}

func TestVoteReq(t *testing.T){
	sm := StateMachine{state: "follower", id:1, peers:[]int{1,2,3,4,7}, term:3, prevLogTerm:2, index:3, leaderid:7, commitIndex:2, leaderCommit:2, lastApplied:2, nextIndex: map[int]int{1:3,2:3,3:3,4:3,5:3} }
	sm.votedFor = 0
	sm.logs[0] = "jmp"
	sm.logs[1] = "jmpz"
	sm.terms[0] = 2
	sm.terms[1] = 2
	sm.lognumber[0] = 1
	sm.lognumber[1] = 2
	var action []Action
	numevents := 0
	action = sm.ProcessEvent(VoteReq{7,3,2,2})
	for i :=0;i<len(action);i++{
		switch action[i].(type){
			case Send:
				numevents++
		}	
	}
	expect(t,string(len(action)),string(1))
	expect(t,string(numevents),string(1))
	
}

func TestVoteResp(t *testing.T){
	sm := StateMachine{state: "candidate", id:1, peers:[]int{1,2,3,4,5}, term:3, prevLogTerm:2, index:3, leaderid:3, commitIndex:2, leaderCommit:2, lastApplied:2, nextIndex: map[int]int{1:3,2:3,3:3,4:3,5:3}, matchIndex: map[int]int{1:2,2:2,3:2,4:2,5:2}}
	sm.logs[0]="jmp"
	sm.logs[1]="jmpz"
	sm.terms[0]=2
	sm.terms[1]=2
	sm.lognumber[0]=1
	sm.lognumber[1]=2
	action := sm.ProcessEvent(VoteResp{2,3,true})
	expect(t,string(len(action)),string(1))	
}

func expect(t *testing.T, a string, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
