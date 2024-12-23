package main 

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"bully_algo/node"
)

func main() {
	fmt.Println("Starting the nodes ......")

	var nodes []*node.Node

	// Create the nodes
	for i := 0; i < 8; i++ {
		n := node.NewNode(i)
		nodes = append(nodes, n)
	}

	// Set the nodes for each node and set the default leader
	for _, n := range nodes {
		n.SetNodes(nodes)
		n.CurrentLeader = 7
		go n.Run()
	}

	// Wait for a termination signal to stop the cluster
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Manually take down node 7 after 10 seconds to trigger the election
	go func() {
		time.Sleep(10 * time.Second)
		nodes[7].Stop()
	}()

	<-sigChan

	fmt.Println("Stopping the nodes ......")
}

// todo : starting and ending of election not being correctly implemented . Fault in design

