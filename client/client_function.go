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

const (
	clientCert = "client/certs/client.pem"
	clientKey  = "client/certs/client.key"
	serverCert = "client/certs/server.pem"
	serverKey  = "client/certs/server.key"

	clientCert1 = "client/certs/client1.pem"
	clientKey1  = "client/certs/client1.key"
	serverCert1 = "client/certs/server1.pem"
	serverKey1  = "client/certs/server1.key"
)

func tcpClient(target string, data []byte) {
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

func tlsClient(target string, server bool) {
	log.Println("Starting tlsClient..")
	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
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

	message := "Hello\n"
	n, err := io.WriteString(conn, message)
	if err != nil {
		log.Fatalf("client: write: %s", err)
	}
	log.Printf("client: wrote %q (%d bytes)", message, n)

	reply := make([]byte, 256)
	n, err = conn.Read(reply)
	log.Printf("client: read %q (%d bytes)", string(reply[:n]), n)

	log.Printf("Upgrading..")
	if server {
		_, err = upgradeToTLSServer(conn.NetConn(), false)
	} else {
		_, err = upgradeToTLSClient(conn.NetConn(), false)
	}
	if err != nil {
		return
	}
}

func dumpConnState(conn *tls.Conn) {
	state := conn.ConnectionState()
	for _, v := range state.PeerCertificates {
		fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
		fmt.Println(v.Subject)
	}
	log.Println("client: handshake: ", state.HandshakeComplete)
}
func handleTLSDispatch(conn *tls.Conn, data []byte) {
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

func handleDispatch(conn net.Conn, data []byte) {

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

func upgradeToTLSServer(conn net.Conn, first bool) (*tls.Conn, error) {
	fmt.Printf("Upgrading the connection to TLS Server\n")
	var cert tls.Certificate
	var err error
	if first {
		fmt.Printf("Loading %s, %s", serverCert, serverKey)
		cert, err = tls.LoadX509KeyPair(serverCert, serverKey)
		if err != nil {
			log.Fatalf("server: loadkeys: %s", err)
			return nil, err
		}
	} else {
		fmt.Printf("Loading %s, %s", serverCert1, serverKey1)
		cert, err = tls.LoadX509KeyPair(serverCert1, serverKey1)
		if err != nil {
			log.Fatalf("server: loadkeys: %s", err)
			return nil, err
		}
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
	dumpConnState(tlsConn)
	return tlsConn, nil
}

func upgradeToTLSClient(conn net.Conn, first bool) (*tls.Conn, error) {
	fmt.Printf("Upgrading the connection to TLS Client\n")
	var cert tls.Certificate
	var err error
	if first {
		fmt.Printf("Loading %s, %s", clientCert, clientKey)
		cert, err = tls.LoadX509KeyPair(clientCert, clientKey)
		if err != nil {
			log.Fatalf("client: loadkeys: %s", err)
			return nil, err
		}
	} else {
		fmt.Printf("Loading %s, %s", clientCert1, clientKey1)
		cert, err = tls.LoadX509KeyPair(clientCert1, clientKey1)
		if err != nil {
			log.Fatalf("server: loadkeys: %s", err)
			return nil, err
		}
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
	dumpConnState(tlsConn)
	return tlsConn, nil
}

func upgradeAgainTLS(nodeConn net.Conn, server bool, data []byte) {
	fmt.Printf("Connected to %s:%s \n", nodeConn.LocalAddr().String(), nodeConn.RemoteAddr().String())
	var tlsConn *tls.Conn
	var err error
	if server {
		tlsConn, err = upgradeToTLSServer(nodeConn, false)
	} else {
		tlsConn, err = upgradeToTLSClient(nodeConn, false)
	}
	if err != nil {
		return
	}
	handleTLSDispatch(tlsConn, data)
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
		tlsConn, err = upgradeToTLSServer(nodeConn, true)
	} else {
		tlsConn, err = upgradeToTLSClient(nodeConn, true)
	}
	if err != nil {
		return
	}
	handleTLSDispatch(tlsConn, data)
	//fmt.Printf("Downgrading the connection to TCP")

	//upgradeAgainTLS(nodeConn, server, data)
	//handleDispatch(nodeConn, data)
}

func tlsServer(target string) {
	cert, err := tls.LoadX509KeyPair(serverCert, serverKey)
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

	// A flock TLS Client which uses tls.Dial to connect to the server's target and then connects via tls to another tls server
	if mode == "flock_tls_client" {
		tlsClient(ipport, false)
		return
	}

	// A regular TLS Server which uses tls.listen to listen to incoming tls connections and then connects via tls to another client
	if mode == "flock_tls_server" {
		tlsClient(ipport, true)
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
	tcpClient(ipport, []byte(message))
}
