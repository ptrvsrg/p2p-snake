package util

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoimpl"

	"p2p-snake/internal/log"
)

const MessageBufferSize = 9192

func SendProto(protoMsg proto.Message, conn *net.UDPConn, daddr *net.UDPAddr) error {
	msg, err := proto.Marshal(protoMsg)
	if err != nil {
		return fmt.Errorf("marshalling error: %v", err)
	}

	_, err = conn.WriteToUDP(msg, daddr)
	if err != nil {
		return fmt.Errorf("sending error: %v\n %v", err, protoimpl.X.MessageStringOf(protoMsg))
	}

	log.Logger.Debugf("sent proto: %v, to: %v", protoimpl.X.MessageStringOf(protoMsg), daddr)
	return nil
}

func ReceiveProto(protoMsg proto.Message, conn *net.UDPConn) (*net.UDPAddr, error) {
	buf := make([]byte, MessageBufferSize)

	err := conn.SetReadDeadline(time.Now().Add(time.Second))
	if err != nil {
		return nil, fmt.Errorf("timeout setting error: %v", err)
	}

	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, fmt.Errorf("receiving error: %v", err)
	}

	err = proto.Unmarshal(buf[:n], protoMsg)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling error: %v", err)
	}

	log.Logger.Debugf("received proto: %v", protoimpl.X.MessageStringOf(protoMsg))
	return addr, nil
}
