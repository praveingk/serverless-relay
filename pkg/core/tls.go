package core

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
)

const (
	serverCert = "client/certs/server.pem"
	serverKey  = "client/certs/server.key"
)

func (r *Relay) tcpServer() error {
	clog.Info("Starting relay... Listening to ", r.url, " for connections")
	acceptor, err := net.Listen("tcp", r.url)
	if err != nil {
		clog.Errorln("Error:", err)
		return err
	}
	var conn1, conn2 net.Conn
	conns := 0
	for {
		ac, err := acceptor.Accept()
		if err != nil {
			clog.Errorln("Accept error:", err)
			continue
		}
		clog.Info("Accept incoming connection from ", ac.RemoteAddr().String())
		conns++
		switch conns {
		case 1:
			conn1 = ac
		case 2:
			conn2 = ac
		}
		if conns == 2 {
			conns = 0
			go r.startForwarding(conn1, conn2)
		}
	}
}

func (r *Relay) tlsServer() error {
	fmt.Printf("Loading %s, %s", serverCert, serverKey)
	cert, err := tls.LoadX509KeyPair(serverCert, serverKey)
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
		return err
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader
	listener, err := tls.Listen("tcp", r.url, &config)
	if err != nil {
		log.Fatalf("server: listen: %s", err)
		return err
	}
	log.Print("server: listening")
	var conn1, conn2 net.Conn
	conns := 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			continue
		}
		defer conn.Close()
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		tlscon, ok := conn.(*tls.Conn)
		if ok {
			state := tlscon.ConnectionState()
			for _, v := range state.PeerCertificates {
				log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
			}
		}
		//go handleClient(conn)

		conns++
		switch conns {
		case 1:
			conn1 = tlscon.NetConn()
			log.Printf("Setting Conn1")
		case 2:
			conn2 = tlscon.NetConn()
			log.Printf("Setting Conn2")
		}
		handleClient(tlscon)
		if conns == 2 {
			conns = 0
			log.Printf("Start forwarding..")
			go r.startForwarding(conn1, conn2)
		}
	}
}

func handleClient(conn net.Conn) {
	buf := make([]byte, 512)
	log.Print("server: conn: waiting")
	n, err := conn.Read(buf)
	if err != nil {
		if err != nil {
			log.Printf("server: conn: read: %s", err)
		}
		return
	}
	log.Printf("server: conn: echo %q\n", string(buf[:n]))
	n, err = conn.Write(buf[:n])

	n, err = conn.Write(buf[:n])
	log.Printf("server: conn: wrote %d bytes", n)

	if err != nil {
		log.Printf("server: write: %s", err)
		return
	}
	log.Println("server: conn: closed")
}
