package gobt

import (
	"io"
	"net"
	"bytes"
	"fmt"
	"time"
)

const (
	ProtocolName = "BitTorrent protocol"
	ProtocolLength = len(ProtocolName)
	HandshakeSize = 49 + ProtocolLength
	HandshakeTimeout = 5 * time.Second
)

// NOTE: Does not support byte flags
func Handshake(conn net.Conn, hash [20]byte, peer [20]byte) error {
	var sendBuf bytes.Buffer
	
	reserved := [8]byte{}

	sendBuf.WriteByte(byte(ProtocolLength))
	sendBuf.WriteString(ProtocolName)
	sendBuf.Write(reserved[:])
	sendBuf.Write(hash[:])
	sendBuf.Write(peer[:])
	
	// Go deadline is a fixed point in time which needs to
	// be always set manually after connection action
	conn.SetDeadline(time.Now().Add(HandshakeTimeout))
	defer conn.SetDeadline(time.Time{})

	if _, err := conn.Write(sendBuf.Bytes()); err != nil {
		return err 
	}

	recvBuf := [HandshakeSize]byte{}

	if _, err := io.ReadFull(conn, recvBuf[:]); err != nil {
		return err
	}
	
	// Return error on incorrect protocol.
	pstrlen := int(recvBuf[0])
	if pstrlen != ProtocolLength {
		return fmt.Errorf("Unexpected protocol identifier length: %d", pstrlen)
	}
	
	pstr := string(recvBuf[1:ProtocolLength+1])
	if pstr != ProtocolName {
		return fmt.Errorf("Unexpected protocol identifier name: %s", pstr)
	}
	
	// Return error on incorrect hash
	recvHash := recvBuf[ProtocolLength+9:ProtocolLength+29]
	if bytes.Compare(hash[:], recvHash) != 0 {
		return fmt.Errorf("Unexpected info hash value: %v", hash)
	}

	return nil
}
