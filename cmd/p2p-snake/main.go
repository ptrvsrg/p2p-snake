package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"p2p-snake/internal/api"
	"p2p-snake/internal/clparser"
	"p2p-snake/internal/config"
	"p2p-snake/internal/hub"
	"p2p-snake/internal/log"
	"p2p-snake/internal/p2p"
)

const title = "                                                                              \n" +
	"    ____ ___   ____       _____             __           _   __          __   \n" +
	"   / __ \\__ \\ / __ \\     / ___/____  ____ _/ /_____     / | / /___  ____/ /__ \n" +
	"  / /_/ /_/ // /_/ /_____\\__ \\/ __ \\/ __ `/ //_/ _ \\   /  |/ / __ \\/ __  / _ \\\n" +
	" / ____/ __// ____/_____/__/ / / / / /_/ / ,< /  __/  / /|  / /_/ / /_/ /  __/\n" +
	"/_/   /____/_/         /____/_/ /_/\\__,_/_/|_|\\___/  /_/ |_/\\____/\\__,_/\\___/ \n" +
	"                                                                              \n" +
	"P2P-Snake-Peer: 1.0.0                                                         \n "

func main() {
	configPath, visible := clparser.Parse()

	fmt.Println(title)
	log.Logger.Info("to quit application press Ctrl+C")

	config.LoadConfig(configPath)

	// Обработчик сигнала SigInt
	sigInt := make(chan os.Signal, 1)
	signal.Notify(sigInt, os.Interrupt, syscall.SIGINT)

	// Инициализация P2P узла
	p2pMulticastAddr, err := net.ResolveUDPAddr("udp",
		net.JoinHostPort(config.Config.P2P.Multicast.Address,
			strconv.Itoa(config.Config.P2P.Multicast.Port)))
	if err != nil {
		log.Logger.Fatalf("resolving P2P node multicast address error: %v", err)
		return
	}
	peer := p2p.NewPeer(p2pMulticastAddr)
	defer func() {
		err := peer.Close()
		if err != nil {
			log.Logger.Error(err)
		}
		log.Logger.Info("P2P node has completed")
	}()

	// Init and start API server
	apiServer := api.NewServer(
		config.Config.API.Port,
		time.Duration(config.Config.API.Timeout)*time.Millisecond,
		peer)
	if err := apiServer.Start(); err != nil {
		log.Logger.Fatal(err)
	}
	defer func() {
		err := apiServer.Close()
		if err != nil {
			log.Logger.Error(err)
		}
		log.Logger.Info("API server has completed")
	}()

	if visible {
		// Init and start of message distribution by free and public nodes
		addrStr := net.JoinHostPort(
			config.Config.Hub.Multicast.Address,
			strconv.Itoa(config.Config.Hub.Multicast.Port),
		)
		addr, err := net.ResolveUDPAddr("udp", addrStr)
		if err != nil {
			log.Logger.Fatalf("hub sender error: %v", err)
		}
		hubSender, err := hub.NewSender(
			config.Config.API.PublicUrl,
			addr,
			func() bool {
				return apiServer.IsFree()
			})
		if err != nil {
			log.Logger.Fatalf("hub sender error: %v", err)
		}

		if err := hubSender.Start(); err != nil {
			log.Logger.Fatalf("hub sender error: %v", err)
		}
		defer func() {
			err := hubSender.Close()
			if err != nil {
				log.Logger.Error(err)
			}
			log.Logger.Info("hub sender has completed")
		}()
	}

	<-sigInt

	log.Logger.Info("waiting for the application to complete")
}
