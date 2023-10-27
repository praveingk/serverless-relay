package core

import (
	"io"
	"net"
	"sync"

	"github.com/allegro/bigcache/queue"
	"github.com/clusterlink-net/clusterlink/pkg/util"
	"github.com/sirupsen/logrus"
)

var clog = logrus.WithField("component", "Frelay-core")
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
func (r *Relay) StartRelay(parsedCertData *util.ParsedCertData) error {
	err := r.tlsServer()
	return err
}

func (r *Relay) startForwarding(conn1, conn2 net.Conn) {
	forwarder := newForwarder(conn1, conn2)
	forwarder.run()
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
func (r *Relay) Init(ip, port string, loglevel logrus.Level) {
	r.url = ip + ":" + port
	clog.Info("Initializing relay for target ", r.target)
	clog.Logger.SetLevel(loglevel)
}
