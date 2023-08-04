/*
 * This file is for temporary testing purpose. TODO : Remove this in future
 */
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

var (
	ctype   = flag.String("type", "server", "server/client")
	ipport  = flag.String("ipport", "9999", "TCP port to listen for flows")
	message = flag.String("message", "test message", "message")
)

var (
	maxDataBufferSize = 64 * 1024
)

func testrecvServiceData(clusterIn string) {
	acceptor, err := net.Listen("tcp", clusterIn)
	if err != nil {
		return
	}
	fmt.Printf("Waiting for connection at %s \n", clusterIn)
	// loop until signalled to stop
	for {
		ac, _ := acceptor.Accept()
		fmt.Printf("Accept connection %s->%s \n", ac.RemoteAddr().String(), ac.LocalAddr().String())
		go recvServiceData(ac, true)
	}
}

func testsendServiceData(clusterIn string, data []byte) {
	nodeConn, err := net.Dial("tcp", clusterIn)
	if err != nil {
		log.Fatalf("Failed to connect to socket %+v", err)
	}
	fmt.Printf("Connected to %s:%s \n", nodeConn.LocalAddr().String(), nodeConn.RemoteAddr().String())
	go recvServiceData(nodeConn, false)
	var i int64
	i = 0
	for {
		nData := strconv.AppendInt(data, i, 10)
		nodeConn.Write(nData)
		time.Sleep(1 * time.Second)
		i++
	}
}

func recvServiceData(conn net.Conn, write bool) {
	bufData := make([]byte, maxDataBufferSize)
	for {
		numBytes, err := conn.Read(bufData)
		if err != nil {
			if err == io.EOF {
				err = nil //Ignore EOF error
			} else {
				log.Fatalf("Read error %v\n", err)
			}
			break
		}
		log.Printf("Received \"%s\" in Socket Connection", bufData[:numBytes])
		if write {
			conn.Write([]byte("Success from server"))
		}
	}
}

func main() {
	flag.Parse()

	switch *ctype {
	case "server":
		testrecvServiceData(*ipport)
	case "client":
		testsendServiceData(*ipport, []byte(*message))
	default:
		fmt.Printf("Wrong Type!")
	}
}
