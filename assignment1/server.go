package main

import(
"net"
"os"
"strings"
"strconv"
"sync"
"time"
"log"
"fmt"
)


var counter int64 = 0
var i int64
type Filecontent struct{
	version int64
	content_len int64
	content string
	starttime time.Time
	expiryTime int64
}
var FileContentLock sync.RWMutex
var file map[string]*Filecontent

func parse(command string) int64{
	//f,_ := os.Create("logs.txt")
	//log.SetOutput(f)
	log.Printf("\nCommandbbb - %s\n",command)
	log.Printf("**************************************************************\n\n")
	arguments := strings.Split(command," ")
	if len(arguments)==4 && arguments[0]=="write"{
		if _, err := strconv.ParseInt(arguments[2],10,64); err != nil {
    		return 5
		}else{
			if _, err := strconv.ParseInt(arguments[3],10,64); err != nil {
    			return 5
			}else{
				return 1
			}
		}
	}else if len(arguments)==3 && arguments[0]=="write"{
		if _, err := strconv.ParseInt(arguments[2],10,64); err != nil {
    		return 5
		}else{
				return 1
			}
	}else if len(arguments)==2 && arguments[0]=="read"{
		return 2
	}else if len(arguments)==2 && arguments[0]=="delete"{
		return 4
	}else if len(arguments)==5  && arguments[0]=="cas"{
		if _, err := strconv.ParseInt(arguments[2],10,64); err != nil {
    		return 5
		}else{
			if _, err := strconv.ParseInt(arguments[3],10,64); err != nil {
    			return 5
			}else{
				if _, err := strconv.ParseInt(arguments[4],10,64); err != nil {
    			return 5
				}else{
					return 3
				}	
			}
		}
	}else if len(arguments)==4  && arguments[0]=="cas"{
		if _, err := strconv.ParseInt(arguments[2],10,64); err != nil {
    		return 5
		}else{
			if _, err := strconv.ParseInt(arguments[3],10,64); err != nil {
    			return 5
			}else{
				return 3
			}
		}
	}else{
		return 5
	}
}

func handleConnection(conn net.Conn){
	f,_ := os.Create("logs.txt")
	log.SetOutput(f)	
	buffer := make([]byte,1)
	defer conn.Close();
	var flag int64 = 0
	var query int64
	var command string = ""
	for {
		n,_ := conn.Read(buffer)
		if n==0 {
			continue
		}
		temp := buffer[0]
		if temp=='\r' {
			flag=1
			continue
		} else if temp=='\n' && flag==1 {
			flag=0
			query = parse(command)
			if query==1 {
				FileContentLock.Lock()
				log.Printf("write command\n")
				arguments := strings.Split(command," ")
				_,ok := file[string(arguments[1])]
				if ok {
					log.Printf("The file laready exists\n")
					//old_version := file[arguments[1]].version
					delete(file,arguments[1])
					var newfile *Filecontent
					newfile = new(Filecontent)
					numbytes,_ :=  strconv.ParseInt(arguments[2],10,64)
					newfile.content_len = int64(numbytes)
					newfile.version = counter
					if len(arguments)==4{
						expTime,_ := strconv.ParseInt(arguments[3],10,64)
						if expTime != 0{
							newfile.expiryTime = int64(expTime)
						}else{
							newfile.expiryTime = -1
						}
					}
					if(len(arguments)==3){
						newfile.expiryTime=-1
					}
					newfile.starttime = time.Now() 
					counter = counter+1
					//file[arguments[1]] = newfile
					var temp1 string = ""
					for i =0;i<numbytes;i++ {
						data,_ :=conn.Read(buffer)
						if data==300{

						}
						character := string(buffer[0])
						temp1 = temp1 + character 
					}
					newfile.content=temp1
					file[arguments[1]]=newfile
					data1,_ :=conn.Read(buffer)
					data2,_ :=conn.Read(buffer)
					if data1 == '\r' && data2 != '\n'{
						log.Printf("data writing not padded properly(no \r\n)")
					}else{
						log.Printf("data writing succesfull\n")
						fmt.Fprint(conn,"OK ",newfile.version,"\r\n")	
					}
				}else{
					log.Printf("creating a new file\n")
					log.Printf(string(arguments[1]))
					log.Printf("\n")
					newfile := new(Filecontent)
					numbytes,_ :=  strconv.ParseInt(arguments[2],10,64)
					newfile.content_len = int64(numbytes)
					//newfile.content =  make([]byte,numbytes)
					newfile.version = counter
					counter = counter+1
					newfile.expiryTime=-1
					if len(arguments)==4{
						expTime,_ := strconv.ParseInt(arguments[3],10,64)
						if expTime != 0{
							newfile.expiryTime = int64(expTime)
						}else{
							newfile.expiryTime = -1
						}
					}
					newfile.starttime = time.Now() 
					//file[arguments[1]] = newfile
					var temp1 string = ""
					for i=0;i<numbytes;i++ {
						data,_ :=conn.Read(buffer)
						if data==300{
							
						}
						character := string(buffer[0])
						temp1 = temp1 + character 
					}
					newfile.content=temp1
					//file[arguments[1]]=newfile
					file[arguments[1]]=newfile
					data1,_ :=conn.Read(buffer)
					data2,_ :=conn.Read(buffer)
					if data1 == '\r' && data2 != '\n'{
						temp = '\r'
						log.Printf("%d %d",data1,temp)
						log.Printf("data writing not padded properly(no \r\n)")
					}else{
						log.Printf("data writing succesfull\n")
						fmt.Fprint(conn,"OK ",newfile.version,"\r\n")

					}
				}
				FileContentLock.Unlock()
			}else if query==2 {
				FileContentLock.RLock()
				log.Printf("read command\n")
				arguments := strings.Split(command," ")
				_,ok := file[arguments[1]]
				if ok{
					d := time.Now().Sub(file[arguments[1]].starttime)
					timeelapsed := int64(d.Seconds())
					if (timeelapsed > file[arguments[1]].expiryTime && file[arguments[1]].expiryTime!=-1) {
						log.Printf("time exceeded\n")
						fmt.Fprint(conn,"ERR_FILE_NOT_FOUND\r\n")
					}else{
					log.Printf("File reading successful\n")
					log.Print(string(file[arguments[1]].content))
					fmt.Fprint(conn,"CONTENTS ",file[arguments[1]].version," ",file[arguments[1]].content_len," ",file[arguments[1]].expiryTime-timeelapsed ,"\r\n")
					fmt.Fprint(conn,file[arguments[1]].content,"\r\n")
					}
				}else{
					log.Printf("File does not exist\n")
					fmt.Fprint(conn,"ERR_FILE_NOT_FOUND\r\n")
				}
				FileContentLock.RUnlock()
			}else if query==3 {
				log.Printf("cas command\n")
				FileContentLock.Lock()
				
				//log.Printf("write command\n")
				arguments := strings.Split(command," ")
				_,ok := file[string(arguments[1])]
				if ok{

					ver,_ := strconv.ParseInt(arguments[2],10,64)
					d := time.Now().Sub(file[arguments[1]].starttime)
					timeelapsed := int64(d.Seconds())
					if (timeelapsed > file[arguments[1]].expiryTime && file[arguments[1]].expiryTime!=-1) {
						log.Printf("time exceeded\n")
						fmt.Fprint(conn,"ERR_FILE_NOT_FOUND\r\n")
						}else if ver == file[arguments[1]].version{
						delete(file,arguments[1])
						var newfile *Filecontent
						newfile = new(Filecontent)
						numbytes,_ :=  strconv.ParseInt(arguments[3],10,64)
						newfile.content_len = int64(numbytes)
						newfile.version = counter
						if len(arguments)==5{
							newfile.expiryTime,_ = strconv.ParseInt(arguments[4],10,64)
						}else{
							newfile.expiryTime = -1
						}
						counter = counter+1
						var temp1 string = ""
						for i=0;i<numbytes;i++ {
							conn.Read(buffer)
						
							character := string(buffer[0])
							temp1 = temp1 + character 
						}
						conn.Read(buffer)
						conn.Read(buffer)
						newfile.content=temp1
						file[arguments[1]]=newfile
						fmt.Fprint(conn,"OK ",newfile.version,"\r\n")	
						}else{
							log.Printf("version mismathc\r\n")
							numbytes,_ := strconv.ParseInt(arguments[3],10,64)
							for i=0;i<numbytes+2;i++ {
							conn.Read(buffer) 
							}
							fmt.Fprint(conn,"ERR_VERSION ",file[arguments[1]].version,"\r\n")
						}
				}else{
					log.Printf("File does not exist\n")
					numbytes,_ := strconv.ParseInt(arguments[3],10,64)
					for i=0;i<numbytes+2;i++ {
							conn.Read(buffer)
						}
					fmt.Fprint(conn,"ERR_FILE_NOT_FOUND\r\n")	
				}

				FileContentLock.Unlock()
			}else if query==4 {
				FileContentLock.Lock()
				log.Printf("delete command\n")
				arguments := strings.Split(command," ")
				_,ok := file[arguments[1]]
				if ok{
					delete(file,arguments[1])
					log.Printf("File %s deleted",arguments[1])
					fmt.Fprint(conn,"OK\r\n")
				}else{
					log.Printf("File %s does not exist",arguments[0])
					fmt.Fprint(conn,"ERR_FILE_NOT_FOUND\r\n")
				}
				FileContentLock.Unlock()
			}else {
				log.Printf("Invalid command\n")
				fmt.Fprint(conn,"ERR_CMD_ERR\r\n")
			}
			command=""
		}else{
			command=command+string(temp) 
		}
		//FileContentLock.Unlock()
	}
}



func serverMain(){
	file = make(map[string]*Filecontent)
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Printf("Error in Listening on server")
		os.Exit(1)
		}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error in Connecting on server")
			os.Exit(1)
		}
	go handleConnection(conn)
}

}

func main() {
	serverMain()
}
