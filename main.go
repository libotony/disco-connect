package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/netutil"
	"github.com/pkg/errors"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: disco-connect <node ID>]\n")
		os.Exit(1)
	}

	target, err := discv5.ParseNode(os.Args[1])
	if err != nil {
		fmt.Println(errors.Wrap(err, "parse node"))
		os.Exit(1)
	}

	verbose := false
	if len(os.Args) == 3 && os.Args[2] == "--verbose" {
		verbose = true
	}

	done := make(chan struct{})
	success := false
	search := fmt.Sprintf("from %x@%v", target.ID[:8], &net.UDPAddr{IP: target.IP, Port: int(target.UDP)})
	searchHandler := log.LazyHandler(log.FuncHandler(func(r *log.Record) error {
		entry := string(log.LogfmtFormat().Format(r))
		if strings.Contains(entry, search) {
			if strings.Contains(entry, "-> known") {
				if !success {
					success = true
					if verbose {
						fmt.Println("↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓")
					}
					fmt.Println("Successfully initiated connection with remote!")
					if verbose {
						fmt.Println("↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑")
					}
					close(done)
				}

			}
		}
		return nil
	}))

	logHandler := searchHandler
	if verbose {
		terminalHandler := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(true)))
		terminalHandler.Verbosity(log.LvlTrace)
		logHandler = log.MultiHandler(searchHandler, terminalHandler)
	}
	log.Root().SetHandler(logHandler)

	err = start(target)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, os.Interrupt, syscall.SIGTERM)

	select {
	case <-done:
		os.Exit(0)
	case sig := <-exitSignal:
		fmt.Printf("Received exit signal: %v\n", sig)
		os.Exit(1)
	case <-time.After(30 * time.Second):
		fmt.Println("Timeout: failed to connect to remote")
		os.Exit(1)
	}
}

func start(target *discv5.Node) error {
	key, err := crypto.GenerateKey()
	if err != nil {
		return err
	}

	addr, err := net.ResolveUDPAddr("udp", ":11235")
	if err != nil {
		return errors.Wrap(err, "-addr")
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	var restrictList *netutil.Netlist
	net, err := discv5.ListenUDP(key, conn, conn.LocalAddr().(*net.UDPAddr), "", restrictList)
	if err != nil {
		return errors.Wrap(err, "start discv5")
	}

	net.SetFallbackNodes([]*discv5.Node{target})
	return nil
}
