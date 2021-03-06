package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	edgefarm_network "github.com/edgefarm/train-simulation/demo/common/go/pkg/edgefarm_network"
	siteevent "github.com/edgefarm/train-simulation/demo/common/go/pkg/siteevent"
	eventlistener "github.com/edgefarm/train-simulation/demo/usecase-5/pkg/eventlistener"
	markdown "github.com/edgefarm/train-simulation/demo/usecase-5/pkg/markdown"
	nats "github.com/nats-io/nats.go"
)

var (
	m *markdown.Markdown
)

func handler(msg *nats.Msg) {
	event, err := siteevent.Unmarshal(msg.Data)
	if err != nil {
		fmt.Printf("Unmarshal failed: %s\n", err)
		return
	}
	m.Add([]string{time.Now().Format("01-02-2006 15:04:05"), event.Site, event.Train, event.Event})
	log.Printf("Train %s, Site %s, Event %s\n", event.Train, event.Site, event.Event)

	err = m.Print()
	if err != nil {
		fmt.Printf("Print failed: %s\n", err)
		return
	}
}

func main() {
	var err error
	m, err = markdown.NewMarkdown("web/public/index.md")
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	m.Print()

	exit := make(chan bool)
	network := edgefarm_network.NewNatsConnection()
	err = network.Connect(10)
	if err != nil {
		fmt.Printf("Error connecting to nats: %s\n", err)
		os.Exit(1)
	}
	lis := eventlistener.NewListener(network, configure(), handler)

	// Listen on signal for termination
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		lis.Close()
		exit <- true
		fmt.Println("signal received, exiting")
	}()
	<-exit
	return
}

// configure reads the environment variable SITE that is in the format <site1>,<site2>,<site3>,...
// and returns a slice of strings with the sites.
func configure() []string {
	sites := ""
	if env := os.Getenv("SITES"); len(env) > 0 {
		sites = env
	}
	return strings.Split(sites, ",")
}
