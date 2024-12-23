package node 

import (
    "fmt"
    "sync"
    "time"
)

type Response int 

const (
	OK Response = iota
	Down 
)

// Node represents a single node in the distributed database.
type Node struct {
    ID            int  // this UID will be important for the election process
    Data          map[string]string
    mu            sync.RWMutex
    CurrentLeader int
    nodes         []*Node
    electionLock  sync.Mutex
    leaderLock    sync.Mutex
    electionStarted bool 
    Active        bool
}

// NewNode creates a new Node with the given ID.
func NewNode(id int) *Node {
    return &Node{
        ID:   id,
        Data: make(map[string]string),
        Active: true,
    }
}

// SetNodes sets the list of nodes in the cluster.
func (n *Node) SetNodes(nodes []*Node) {
    n.nodes = nodes
}

// Put stores a key-value pair in the node.
func (n *Node) Put(key, value string) {
    n.mu.Lock()
    defer n.mu.Unlock()
    n.Data[key] = value
}

// Get retrieves the value for a given key from the node.
func (n *Node) Get(key string) (string, bool) {
    n.mu.RLock()
    defer n.mu.RUnlock()
    value, exists := n.Data[key]
    return value, exists
}

// Delete removes a key-value pair from the node.
func (n *Node) Delete(key string) {
    n.mu.Lock()
    defer n.mu.Unlock()
    delete(n.Data, key)
}

// PrintData prints all key-value pairs stored in the node.
func (n *Node) PrintData() {
    n.mu.RLock()
    defer n.mu.RUnlock()
    fmt.Printf("Node %d data:\n", n.ID)
    for key, value := range n.Data {
        fmt.Printf("%s: %s\n", key, value)
    }
}

// Run starts the node and prints a running message.
func (n *Node) Run() {
    fmt.Printf("Node %d is running\n", n.ID)
    if n.ID == n.CurrentLeader {
        fmt.Printf("Node %d is the leader\n", n.ID)
    }
    go n.Ping()
}

// StartElection initiates the leader election process.
func (n *Node) StartElection() {
    n.electionLock.Lock()
    defer n.electionLock.Unlock()

    if n.electionStarted {
        return
    }

    n.electionStarted = true
    fmt.Printf("Node %d is starting an election\n", n.ID)
    go n.selectLeader()
}

// selectLeader selects the new leader.
func (n *Node) selectLeader() {
    n.leaderLock.Lock()
    defer n.leaderLock.Unlock()

    for _, node := range n.nodes {
        if node.ID > n.ID {
            status := node.Challenge(n.ID)
            if status == OK {
                return
            }
        }
    }

    n.DeclareLeader()
    n.electionStarted = false
}

// Challenge handles a challenge message from another node.
func (n *Node) Challenge(challengerID int) Response {
    if n.Active {
		go n.selectLeader()
        return OK
    } else {
        return Down
    }
}

// Ping ends ping messages to nodes with higher IDs in a round-robin fashion.
func (n *Node) Ping() {
    for {
        for _, node := range n.nodes {
            if node.ID > n.ID {
                status := node.Pong(n.ID)
                if status == Down && n.CurrentLeader == node.ID {
                    n.StartElection()
                }
                time.Sleep(2 * time.Second)
            }
        }
    }
}

// Ping handles a ping message from another node.
func (n *Node) Pong(pingerID int) Response {
    if n.Active {
        return OK
    } else {
        return Down
    }
}

// DeclareLeader declares the current node as the new leader.
func (n *Node) DeclareLeader() {
    n.CurrentLeader = n.ID
    fmt.Printf("I, Node %d, am the new leader\n", n.ID)
}

// Stop stops the node and its associated goroutine.
func (n *Node) Stop() {
    n.Active = false
    fmt.Printf("Node %d is stopped\n", n.ID)
}