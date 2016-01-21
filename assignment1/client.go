package main

import(
"fmt"
"net"
"os"
"time"
)

func clientMain() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("Error in Connecting on client")
		os.Exit(1)
	}
	//fmt.Fprintf(conn, "GET / HTTP/1.0\r\n")
	fmt.Fprint(conn,"write surya 17 abc\r\nthe champ is here\r\nread surya\r\nwrite surya 22 5\r\nthe champ should sleep\r\ncjvngf\r\nwrite surya 22\r\nthe champ should sleep\r\n")
	time.Sleep(5*time.Second)
	fmt.Fprintf(conn,"read surya\r\n")
	//fmt.Fprint(conn,"delete surya\r\n")
	buffer := make([]byte,1)
	flag := 1
	message := ""
	for{
		n,err := conn.Read(buffer)
		if err!=nil {
			os.Exit(1)
		}
		temp := string(buffer[0])
		if(n==0){
			if flag==0{
				fmt.Printf(message,"\n")
				flag=1
			}
			continue
		}else{
			flag=0
			fmt.Printf(temp)
			message = message+temp	
		}
	}
	//status, err := bufio.NewReader(conn).ReadString('\n')
}


func main() {
	clientMain()
}
