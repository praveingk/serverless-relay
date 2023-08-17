package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

var (
	maxDataBufferSize = 64 * 1024
)

func tcp_client(target string, data []byte) {
	log.Printf("Connecting to %s", target)
	nodeConn, err := net.Dial("tcp", target)
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
	for i := 0; i < 100; i++ {
		numBytes, err := conn.Read(bufData)
		if err != nil {
			fmt.Printf("Read error %v\n", err)
			break
		}
		fmt.Printf("Received \"%s\"\n", bufData[:numBytes])
		if write {
			conn.Write([]byte("Success from server"))
		}
	}
}

func tlsClient(target string) {
	cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", target, &config)
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}
	defer conn.Close()
	log.Println("client: connected to: ", conn.RemoteAddr())

	state := conn.ConnectionState()
	for _, v := range state.PeerCertificates {
		fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
		fmt.Println(v.Subject)
	}
	log.Println("client: handshake: ", state.HandshakeComplete)
	log.Println("client: mutual: ", state.NegotiatedProtocolIsMutual)

	message := "Hello\n"
	n, err := io.WriteString(conn, message)
	if err != nil {
		log.Fatalf("client: write: %s", err)
	}
	log.Printf("client: wrote %q (%d bytes)", message, n)

	reply := make([]byte, 256)
	n, err = conn.Read(reply)
	log.Printf("client: read %q (%d bytes)", string(reply[:n]), n)
	log.Print("client: exiting")
}

func handleTLSDispatch(conn *tls.Conn, data []byte) {
	state := conn.ConnectionState()
	for _, v := range state.PeerCertificates {
		fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
		fmt.Println(v.Subject)
	}
	log.Println("client: handshake: ", state.HandshakeComplete)
	log.Println("client: mutual: ", state.NegotiatedProtocolIsMutual)

	go recvServiceData(conn, false)
	var i int64
	i = 0
	for {
		nData := strconv.AppendInt(data, i, 10)
		conn.Write(nData)
		time.Sleep(1 * time.Second)
		i++
	}
}

func upgradeToTLSServer(conn net.Conn) (*tls.Conn, error) {
	fmt.Printf("Upgrading the connection to TLS Server\n")

	cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
		return nil, err
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	tlsConn := tls.Server(conn, &config)

	err = tlsConn.Handshake()
	if err != nil {
		log.Fatalf("failed to perform handshake : %+v", err)
		return nil, err
	}
	fmt.Printf("Handshake complete\n")
	time.Sleep(1 * time.Second)
	return tlsConn, nil
}

func upgradeToTLSClient(conn net.Conn) (*tls.Conn, error) {
	fmt.Printf("Upgrading the connection to TLS Client\n")
	cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
	if err != nil {
		log.Fatalf("client: loadkeys: %s", err)
		return nil, err
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	tlsConn := tls.Client(conn, &config)
	err = tlsConn.Handshake()
	if err != nil {
		log.Fatalf("failed to perform handshake : %+v", err)
		return nil, err
	}
	fmt.Printf("Handshake complete\n")
	time.Sleep(1 * time.Second)

	return tlsConn, nil
}

// upgradeTLS connects to a target using a regular net.Dial TCP.
// Upon success, it upgrades to tls server/client based on MODE env variable.
// Further, all read/writes are end-to-end TLS
func upgradeTLS(target string, server bool, data []byte) {
	log.Printf("Connecting to %s", target)
	nodeConn, err := net.Dial("tcp", target)
	if err != nil {
		log.Fatalf("Failed to connect to socket %+v", err)
	}
	fmt.Printf("Connected to %s:%s \n", nodeConn.LocalAddr().String(), nodeConn.RemoteAddr().String())
	var tlsConn *tls.Conn
	if server {
		tlsConn, err = upgradeToTLSServer(nodeConn)
	} else {
		tlsConn, err = upgradeToTLSClient(nodeConn)
	}
	if err != nil {
		return
	}
	handleTLSDispatch(tlsConn, data)
}

func tlsServer(target string) {
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader
	listener, err := tls.Listen("tcp", target, &config)
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}
	log.Print("server: listening")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		defer conn.Close()
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		tlscon, ok := conn.(*tls.Conn)
		if ok {
			log.Print("ok=true")
			state := tlscon.ConnectionState()
			for _, v := range state.PeerCertificates {
				log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
			}
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 512)
	for {
		log.Print("server: conn: waiting")
		n, err := conn.Read(buf)
		if err != nil {
			if err != nil {
				log.Printf("server: conn: read: %s", err)
			}
			break
		}
		log.Printf("server: conn: echo %q\n", string(buf[:n]))
		n, err = conn.Write(buf[:n])

		n, err = conn.Write(buf[:n])
		log.Printf("server: conn: wrote %d bytes", n)

		if err != nil {
			log.Printf("server: write: %s", err)
			break
		}
	}
	log.Println("server: conn: closed")
}

func main() {
	ipport := os.Getenv("TARGET")
	message := os.Getenv("MESSAGE")
	mode := os.Getenv("MODE")

	// A regular TLS Client which uses tls.Dial to connect to the server's target
	if mode == "regular_tls_client" {
		tlsClient(ipport)
		return
	}

	// A regular TLS Server which uses tls.listen to listen to incoming tls connections
	if mode == "regular_tls_server" {
		tlsServer(ipport)
		return
	}

	// A TLS server which connects to the relay using regular TCP, and then upgrades to TLS Server
	if mode == "tls_server" {
		upgradeTLS(ipport, true, []byte(message))
		return
	}

	// A TLS server which connects to the relay using regular TCP, and then upgrades to TLS Server
	if mode == "tls_client" {
		upgradeTLS(ipport, false, []byte(message))
		return
	}
	// Nothing else set, then revert to a normal tcp client
	tcp_client(ipport, []byte(message))
}
