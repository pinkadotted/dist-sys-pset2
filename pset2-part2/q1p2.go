package pset2part2

import (
	"fmt"
	"time"
)

type Node struct {
	id              int
	messageChan     chan Message
	vectorClock     [10]int
	priorityQueue   []Message
	isInCS          bool
	wantToEnterCS   bool
	receivedReplies []int // slice of ids of received replies
	time_requested  time.Time
	time_exited     time.Time
	hasBeenInCS     bool
}

type Message struct {
	senderID          int
	messageType       string // can be "request", "reply", "release"
	senderVectorClock [10]int
}

//	HELPER FUNCTIONS:
//
// func to init a node
func initNode(id int) Node {
	n := Node{}
	n.id = id
	n.messageChan = make(chan Message, 100)
	n.vectorClock[id] = 0
	n.priorityQueue = make([]Message, 0)
	n.receivedReplies = make([]int, 0)
	return n
}

// func to check if a vector clock is less than another
func vectorClockLessThan(smallerMessage, largerMessage *Message) bool {
	hasHigher := false
	hasLower := false
	for i := 0; i < numberOfNodes; i++ {
		if smallerMessage.senderVectorClock[i] > largerMessage.senderVectorClock[i] {
			hasHigher = true
		} else if smallerMessage.senderVectorClock[i] < largerMessage.senderVectorClock[i] {
			hasLower = true
		}
	}
	if hasHigher && hasLower {
		return smallerMessage.senderID < largerMessage.senderID
	} else if hasHigher && !hasLower {
		return false
	}
	return true
}

// func to get Node from id
func getNodeFromID(id int) *Node {
	return &nodes[id]
}

// MAIN FUNCTIONS:
// function to listen for messages
func (n *Node) listenForMessages() {
	for {
		msg := <-n.messageChan
		// fmt.Printf("%d: received message from %d of type %s\n", n.id, msg.senderID, msg.messageType)
		if msg.messageType == "request" {
			go n.replyToRequest(msg)
		} else if msg.messageType == "reply" {
			n.receivedReplies = append(n.receivedReplies, msg.senderID)
			// fmt.Printf("%d RECEIVED REPLIES: %v\n", n.id, n.receivedReplies)
			if len(n.receivedReplies) == numberOfNodes-1 {
				// enter critical section
				go n.enterCriticalSection()
			}
		}
	}
}

// function to update the vector clock of a node
func (n *Node) updateVectorClock(msg Message) {
	for i := 0; i < numberOfNodes; i++ {
		if msg.senderVectorClock[i] > n.vectorClock[i] {
			n.vectorClock[i] = msg.senderVectorClock[i]
		}
	}
}

// function to insert message into priority queue
func (n *Node) insertMessageIntoPQueue(m Message) {
	// fmt.Println("Node", n.id, "inserting message into priority queue...")
	for i, msg := range n.priorityQueue {
		if vectorClockLessThan(&m, &msg) {
			n.priorityQueue = append(n.priorityQueue[:i], append([]Message{m}, n.priorityQueue[i:]...)...)
			// fmt.Printf("Node %d has new priority queue: %v\n", n.id, n.priorityQueue)
			return
		}
	}
	n.priorityQueue = append(n.priorityQueue, m)
	// fmt.Printf("Node %d has new priority queue: %v\n", n.id, n.priorityQueue)
}

// function to send a reply to a node that has requested to enter critical section
func (n *Node) replyToRequest(msg Message) {
	requester := getNodeFromID(msg.senderID)
	myStatus := Message{
		senderID:          n.id,
		messageType:       "reply",
		senderVectorClock: n.vectorClock}
	for {
		if !n.isInCS {
			// if i do not want to enter the critical section, i will send a reply to the node that has requested to enter critical section
			if !n.wantToEnterCS || (n.wantToEnterCS && vectorClockLessThan(&msg, &myStatus)) {
				// fmt.Printf("%d's vector clock: %v, received vector clock: %v\n", n.id, n.vectorClock, msg.senderVectorClock)
				requester.messageChan <- myStatus
				n.updateVectorClock(msg)
				return
			} else if n.wantToEnterCS && vectorClockLessThan(&myStatus, &msg) { // I want to enter CS & I am higher priority than the request I have received
				n.insertMessageIntoPQueue(msg)
				// fmt.Printf("Node %d's pQueue: %v\n", n.id, n.priorityQueue)
				return
			}
		} else {
			n.insertMessageIntoPQueue(msg)
			// fmt.Printf("Node %d's pQueue: %v\n", n.id, n.priorityQueue)
			return
		}
	}
}

// simulate entering critical section by just sleeping for a random amount of time
func (n *Node) enterCriticalSection() {
	n.isInCS = true
	fmt.Println("Node", n.id, "ENTERING critical section...")
	time.Sleep(time.Duration(1) * time.Second)
	n.vectorClock[n.id]++
	n.receivedReplies = make([]int, 0)
	n.wantToEnterCS = false
	n.isInCS = false
	fmt.Println("Node", n.id, "EXITING critical section...")
	for _, message := range n.priorityQueue {
		n.replyToRequest(message)
	}
	n.priorityQueue = make([]Message, 0)
	n.time_exited = time.Now()
}

// function to automatically, periodically send a request to enter critical section to all the other nodes via the message channel
func (n *Node) requestToEnterCriticalSection() {
	// for {
	for !n.hasBeenInCS {
		// time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
		n.time_requested = time.Now()
		if !n.wantToEnterCS {
			n.receivedReplies = make([]int, 0)
			n.wantToEnterCS = true
			n.vectorClock[n.id]++
			VCsnapshot := n.vectorClock
			// fmt.Printf("Node %d requesting to enter critical section... VC: %v\n", n.id, n.vectorClock)
			for i := 0; i < numberOfNodes; i++ {
				if i != n.id {
					getNodeFromID(i).messageChan <- Message{
						senderID:          n.id,
						messageType:       "request",
						senderVectorClock: VCsnapshot,
					}
				}
			}
		}
		n.hasBeenInCS = true
	}
}

var nodes []Node
var numberOfNodes = 10

// make a map of node ids to their times in CS
var timesInCS = make(map[int]time.Duration)

// function to initialize all the nodes, and start the listening for messages and request to enter critical section goroutines, and randomly send requests to enter critical section
func Q1P2() {
	fmt.Println("Q1P2 starting...")
	for i := 0; i < numberOfNodes; i++ {
		nodes = append(nodes, initNode(i))
	}
	for i := 0; i < numberOfNodes; i++ {
		go nodes[i].listenForMessages()
		go nodes[i].requestToEnterCriticalSection()
	}
	// i want to check if all the timesInCS have been filled
	for len(timesInCS) < numberOfNodes {
		// i want to check if a node has exited the critical section
		for i := 0; i < numberOfNodes; i++ {
			if (nodes[i].time_exited.Sub(nodes[i].time_requested)) > time.Duration(0) && timesInCS[i] == 0 {
				timesInCS[i] = nodes[i].time_exited.Sub(nodes[i].time_requested)
				// i want to print all times
				fmt.Printf("Node %d's time in CS: %v\n", i, nodes[i].time_exited.Sub(nodes[i].time_requested))
				fmt.Printf("Length of timesInCS: %d\n", len(timesInCS))
			}
		}
	}
}
