package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)


// Simple serial check of getting and setting
func TestTCPSimple(t *testing.T) {
	go serverMain()
        time.Sleep(1 * time.Second) // one second is enough time for the server to start
	name := "hi.txt"
	contents := "bye"
	exptime := 300000
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Error(err.Error()) // report error through testing framework
	}

	scanner := bufio.NewScanner(conn)

	// Write a file
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan() // read first line
	resp := scanner.Text() // extract the text from the buffer
	arr := strings.Split(resp, " ") // split into OK and <version>
	expect(t, arr[0], "OK")
	version, err := strconv.ParseInt(arr[1], 10, 64) // parse version as number
	if err != nil {
		t.Error("Non-numeric version found")
	}
	time.Sleep(5*time.Second)
	fmt.Fprintf(conn, "read %v\r\n", name) // try a read now
	scanner.Scan()

	arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "CONTENTS")
	expect(t, arr[1], fmt.Sprintf("%v", version)) // expect only accepts strings, convert int version to string
	expect(t, arr[2], fmt.Sprintf("%v", len(contents)))
	expect(t, arr[3], fmt.Sprintf("%v",exptime-5))
	scanner.Scan()
	expect(t, contents, scanner.Text())

	//Reading a nonexisting file
	fmt.Fprintf(conn, "read notexist\r\n")
	scanner.Scan()
	expect(t, "ERR_FILE_NOT_FOUND", scanner.Text())

	//comparing and swaping a file	
	contents = "message number 2"
	fmt.Fprintf(conn, "cas %v %v %v\r\n%v\r\n",name,version,len(contents),contents)
	//fmt.Fprintf(conn, contents,"\r\n")
	scanner.Scan()
	resp = scanner.Text() // extract the text from the buffer
	arr = strings.Split(resp," ")
	expect(t,arr[0],"OK")
	version,_ = strconv.ParseInt(arr[1],10,64)


	//comparing and swapping a non existent file
	fmt.Fprintf(conn, "cas nonexisting %v %v\r\n%v\r\n",version,len(contents),contents)
	scanner.Scan()
	expect(t, "ERR_FILE_NOT_FOUND",scanner.Text())

	//version mismatch in cas
	fmt.Fprintf(conn, "cas %v %v %v\r\n%v\r\n",name,version-10,len(contents),contents)
	scanner.Scan()
	resp = scanner.Text()
	arr = strings.Split(resp," ")
	expect(t, "ERR_VERSION",arr[0])	
	




	//delete an existing file
	fmt.Fprintf(conn, "delete %v\r\n", name)
	scanner.Scan()
	expect(t,"OK",scanner.Text())

	//read from a deleted file
	fmt.Fprintf(conn,"read %v\r\n",name)
	scanner.Scan()
	expect(t,"ERR_FILE_NOT_FOUND",scanner.Text())


}

// Useful testing function
func expect(t *testing.T, a string, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}