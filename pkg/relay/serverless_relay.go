package relay

import (
	"io"
	"net"
	"strings"
	"sync"

	"github.com/allegro/bigcache/queue"
	"github.com/sirupsen/logrus"
)

var clog = logrus.WithField("component", "Serverless-Relay")
var queueSize = 100

var localhost = "127.0.0.1"
var (
	maxDataBufferSize = 64 * 1024
)

// Relay struct defines the properties of the relay
type Relay struct {
	gatewayConn    net.Conn
	clientConn     net.Conn
	egress         queue.BytesQueue // Message queued to be sent to gateway
	ingress        queue.BytesQueue // Messages queued to be sent to client
	egressDrain    sync.Mutex
	ingressDrain   sync.Mutex
	ingressHandler bool
	url            string
	target         string
	gwIP           string
}

// StartRelay starts the main function of the relay
func (r *Relay) StartRelay() error {
	// Start a TCP Server at ip:port
	// Accept only two connections, one from the client, and one from the gateway
	// Distinguish the connections, based on the IP. Assume we know the gateway IP
	clog.Info("Starting relay... Listening to ", r.url, " for connections")
	acceptor, err := net.Listen("tcp", r.url)
	if err != nil {
		clog.Errorln("Error:", err)
		return err
	}
	// loop until signalled to stop
	for {
		ac, err := acceptor.Accept()
		if err != nil {
			clog.Errorln("Accept error:", err)
			continue
		}
		clog.Info("Accept incoming connection from ", ac.RemoteAddr().String())
		addr := strings.Split(ac.RemoteAddr().String(), ":")[0]
		clog.Debug("Comparing ", addr, " and ", r.gwIP)
		if addr == r.gwIP || addr == localhost {
			// This is an incoming connection from gateway
			if r.gatewayConn != nil {
				clog.Errorln("Preexisting gateway connection still active.")
				ac.Close()
				continue
			}
			clog.Info("Got a connection from the gateway")
			r.gatewayConn = ac
			// Check messages queued
			go r.drainEgress()
			if !r.ingressHandler {
				go r.handleIngress()
			}
		} else {
			// This is an incoming connection from a client
			if r.clientConn != nil {
				clog.Errorln("Preexisting client connection still active.")
				ac.Close()
				continue
			}
			clog.Info("Got a connection from the client")
			r.clientConn = ac
			go r.drainIngress()
			if r.gatewayConn == nil {
				// If no gateway connection open yet, try reaching target
				clog.Debug("Trying to reach the target..")
				conn, err := net.Dial("tcp", r.target)
				if err != nil {
					clog.Errorln("Unable to reach target, will be buffering")
				}
				r.gatewayConn = conn
			}
			go r.handleEgress()
			if !r.ingressHandler {
				go r.handleIngress()
			}
		}
	}
	return nil
}

// Starts to emit the messages destined to the gateway
func (r *Relay) drainEgress() error {
	r.egressDrain.Lock()
	clog.Debug("Locking Egress Drain from drainEgress")

	var err error
	for {
		// Dequeue messages from Egress
		message, err := r.egress.Pop()
		if err != nil {
			clog.Debug("Unlocking Egress Drain from drainEgress")
			r.egressDrain.Unlock()
			return err
		}
		// Send to gatewayConn
		_, err = r.gatewayConn.Write(message)
		if err != nil {
			clog.Debugf("Drain Egress: Write error %v\n", err)
			clog.Debug("Unlocking Egress Drain from drainEgress")
			r.egressDrain.Unlock()
			break
		}
	}
	clog.Debugf("Unlocking egress Drain")
	r.egressDrain.Unlock()
	return err
}

// Starts to emit the messages destined to the client
func (r *Relay) drainIngress() error {
	r.ingressDrain.Lock()
	clog.Debug("Locking Ingress Drain from drainIngress")
	var err error
	for {
		// Dequeue messages from Ingress
		message, err := r.ingress.Pop()
		if err != nil {
			clog.Debug("Unlocking Ingress Drain from drainIngress")
			r.ingressDrain.Unlock()
			return err
		}
		// Send to clientConn
		_, err = r.clientConn.Write(message)
		if err != nil {
			clog.Debugf("Drain Ingress: Write error %v\n", err)
			clog.Debug("Unlocking Ingress Drain from drainIngress")
			r.ingressDrain.Unlock()
			return err
		}
	}
	clog.Debug("Unlocking Ingress Drain from drainIngress")

	r.ingressDrain.Unlock()
	return err
}

// Starts to Listen to gateway connection and queue messages/sends them to client connection in ingress
func (r *Relay) handleIngress() error {
	r.ingressHandler = true
	var err error
	bufData := make([]byte, maxDataBufferSize)
	lock := false
	if r.gatewayConn != nil {
		defer r.gatewayConn.Close()
	}
	for {
		numBytes, err := r.gatewayConn.Read(bufData)
		if err != nil {
			clog.Error("Handle Ingress: Read error ", err)
			break
		}
		clog.Debugf("Read from gateway : %s", string(bufData))
		if r.clientConn == nil {
			clog.Debug("Queueing in Ingress")
			_, err := r.ingress.Push(bufData)
			if err != nil {
				clog.Error("Unable to push to ingress : ", err)
			}
		} else {
			if !lock {
				clog.Debug("Locking Ingress Drain from handleIngress")
				r.ingressDrain.Lock()
				lock = true
				clog.Debugf("Length of Ingress Queue : %d : This must be 0 ideally", r.ingress.Len())
			}
			_, err = r.clientConn.Write(bufData[:numBytes])
			if err != nil {
				clog.Error("Handle Ingress: Write error ", err)
				break
			}
			clog.Debug("Finished Writing to client connection")
		}
	}
	if lock {
		clog.Debug("Unlocking Ingress Drain from handleIngress")
		r.ingressDrain.Unlock()
	}
	r.closeConnection()
	r.ingressHandler = false
	if err == io.EOF {
		return nil
	} else {
		return err
	}
}

// Starts to Listen to client connection and queue messages/sends them to gateway connection in egress
func (r *Relay) handleEgress() error {
	var err error
	bufData := make([]byte, maxDataBufferSize)
	lock := false
	if r.clientConn != nil {
		defer r.clientConn.Close()
	}
	for {
		numBytes, err := r.clientConn.Read(bufData)
		if err != nil {
			clog.Error("Handle Egress: Read error ", err)
			break
		}
		clog.Debugf("Read from client : %s", string(bufData))
		if r.gatewayConn == nil {
			clog.Debug("Queueing in Egress")
			_, err := r.egress.Push(bufData)
			if err != nil {
				clog.Error("unable to push to egress : ", err)
			}
		} else {
			if !lock {
				clog.Debug("Locking Egress Drain from handleEgress")
				r.egressDrain.Lock()
				lock = true
				clog.Debugf("Length of Egress Queue : %d : This must be 0 ideally", r.egress.Len())
			}

			_, err = r.gatewayConn.Write(bufData[:numBytes])
			if err != nil {
				clog.Error("Handle Egress: Write error ", err)
				break
			}
			clog.Debug("Finished Writing to gateway connection")
		}
	}
	clog.Infof("Lock held = %+v", lock)
	if lock {
		clog.Debug("Unlocking Egress Drain from handleEgress")
		r.egressDrain.Unlock()
	}
	r.closeConnection()
	if err == io.EOF {
		return nil
	} else {
		return err
	}

}

// Close connections
func (r *Relay) closeConnection() {
	clog.Info("Closing Connections")
	if r.gatewayConn != nil {
		r.gatewayConn.Close()
		r.gatewayConn = nil
	}
	if r.clientConn != nil {
		r.clientConn.Close()
		r.clientConn = nil
	}

}

// Init initializes the relay
func (r *Relay) Init(ip, port, target string, loglevel logrus.Level) {
	r.url = ip + ":" + port
	r.target = target
	r.gwIP = strings.Split(target, ":")[0]
	r.clientConn = nil
	r.gatewayConn = nil
	r.egress = *queue.NewBytesQueue(0, queueSize*maxDataBufferSize, false)
	r.ingress = *queue.NewBytesQueue(0, queueSize*maxDataBufferSize, false)
	clog.Info("Initializing relay for target ", r.target)
	clog.Logger.SetLevel(loglevel)
}
