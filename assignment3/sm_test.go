package main 

import(
	
	"fmt"
	"time"
	"testing"
	"strconv"
	"github.com/cs733-iitb/log"
	"github.com/cs733-iitb/cluster"
	"github.com/cs733-iitb/cluster/mock"
)

var NumServers = 5
var ETimeout = 1000
var HeartBeat = 500

func makerafts() ([]*Node){
	  
	newconfig := cluster.Config{
    Peers: []cluster.PeerConfig{
        	{Id: 1, Address: "localhost:5050"},
        	{Id: 2, Address: "localhost:6060"},
            {Id: 3, Address: "localhost:7070"},
            {Id: 4, Address: "localhost:8080"},
            {Id: 5, Address: "localhost:9090"},
        	}}
    cl,_ := mock.NewCluster(newconfig)
     	
    nodes := make([]*Node, len(newconfig.Peers))
    
    TempConfig := ConfigRaft{
    	LogDir: "Log",
    	ElectionTimeout: ETimeout,
    	HeartbeatTimeout: HeartBeat,
    }

    MatchIndex := map[int]int{0:-1,1:-1,2:-1,3:-1,4:-1}
    NextIndex := map[int]int{0:1,1:0,2:0,3:0,4:0}
    Votes  := make([]int, NumServers)
    for i:=0; i<NumServers;i++ { 
    	Votes[i] = 0
    }
   
    for i:=1; i<=NumServers;i++ { 
    	RaftNode := Node{id:i, leaderid:-1, timeoutCh:time.NewTimer(time.Duration(TempConfig.ElectionTimeout)*time.Millisecond), LogDir:TempConfig.LogDir+strconv.Itoa(i), CommitCh:make(chan *Commit,100)}
    	log, err := log.Open(RaftNode.LogDir)
		log.RegisterSampleEntry(LogStore{})
		if err!=nil {
			fmt.Println("Error opening Log File"+"i")
		}
		peer := make([]int,NumServers)
		server := cl.Servers[i]
		p := server.Peers()
		temp := 0

		for j:=0;j<NumServers;j++{ 
			if j!=i-1{ 
				peer[j]=0
			}else{ 
				peer[j]=p[temp]
				temp++
			}
		}

		RaftNode.server = server
		RaftNode.sm = &StateMachine{state:"follower", id: server.Pid(), leaderid:-1, peers:peer, term:0, votedFor:0, Votes:Votes, commitIndex:-1, Log:log, index: -1, matchIndex:MatchIndex, nextIndex: NextIndex,HeartbeatTimeout:500,ElectionTimeout:ETimeout }
		nodes[i-1] = &RaftNode
    }
    for i:=0;i<5;i++{
		//fmt.Println(i) 
		go nodes[i].listenonchannel()
	}

	return nodes
}

func TestBasic(t *testing.T){ 
	makerafts()
	time.Sleep(time.Second*4)

}


