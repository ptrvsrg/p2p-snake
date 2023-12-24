package hub

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"

	"p2p-snake/internal/log"
	"p2p-snake/internal/util"
)

const (
	version = 1
	timeout = time.Second
)

type Sender struct {
	id            string
	publicUrl     string
	multicastAddr *net.UDPAddr
	conn          *net.UDPConn
	canSend       func() bool

	// Close
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func NewSender(publicUrl string, multicastAddr *net.UDPAddr, canSend func() bool) (*Sender, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	return &Sender{
		id:            id.String(),
		publicUrl:     publicUrl,
		multicastAddr: multicastAddr,
		canSend:       canSend,

		cancel: func() {},
		wg:     &sync.WaitGroup{},
	}, nil
}

func (sender *Sender) Start() error {
	var err error
	sender.conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return err
	}
	log.Logger.Infof("hub sender running on %v", sender.conn.LocalAddr())

	var ctx context.Context
	ctx, sender.cancel = context.WithCancel(context.Background())
	sender.wg.Add(1)
	go sender.send(ctx)

	return nil
}

func (sender *Sender) send(ctx context.Context) {
	defer sender.wg.Done()

	msg := NewHubMessage(int32(version), sender.id, sender.publicUrl)
	log.Logger.Debug("send goroutine is running")
	for {
		select {
		case <-ctx.Done():
			log.Logger.Debug("send goroutine has completed")
			return
		case <-time.After(timeout):
			if sender.canSend() {
				err := util.SendProto(msg, sender.conn, sender.multicastAddr)
				if err != nil {
					log.Logger.Errorf("hub sender error: %v", err)
				}
			}
		}
	}
}

func (sender *Sender) Close() error {
	sender.cancel()
	sender.wg.Wait()
	return sender.conn.Close()
}
