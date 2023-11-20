package pset2part1

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

func initNode(id int) Node {
	n := Node{}
	n.id = id
	n.messageChan = make(chan Message, 100)
	n.vectorClock[id] = 1
	n.priorityQueue = make([]Message, 0)
	n.receivedReplies = make([]int, 0)
	return n
}

//	HELPER FUNCTIONS:
//
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

// func to check if ack has been received from requester
func (n *Node) checkIfReceivedAck(requester *Node) bool {
	for _, reply := range n.receivedReplies {
		if reply == requester.id {
			return true
		}
	}
	return false
}

// MAIN FUNCTIONS:
// function to listen for messages
func (n *Node) listenForMessages() {
	for {
		msg := <-n.messageChan
		// fmt.Printf("%d: received message from %d of type %s\n", n.id, msg.senderID, msg.messageType)
		n.updateVectorClock(msg)
		if msg.messageType == "request" {
			n.insertMessageIntoPQueue(msg)
			go n.replyToRequest(msg)
		} else if msg.messageType == "release" {
			if len(n.priorityQueue) <= 1 {
				n.priorityQueue = make([]Message, 0)
			} else {
				n.priorityQueue = n.priorityQueue[1:]
				// fmt.Println("Node", n.id, "has new priority queue after RELEASE MSG:", n.priorityQueue)
			}
			if n.wantToEnterCS && len(n.receivedReplies) == numberOfNodes-1 && len(n.priorityQueue) > 0 && n.priorityQueue[0].senderID == n.id {
				// enter critical section
				go n.enterCriticalSection()
			}
		} else if msg.messageType == "reply" {
			n.receivedReplies = append(n.receivedReplies, msg.senderID)
			// fmt.Printf("%d RECEIVED REPLIES: %v\n", n.id, n.receivedReplies)
			if len(n.receivedReplies) == numberOfNodes-1 && len(n.priorityQueue) > 0 && n.priorityQueue[0].senderID == n.id {
				// enter critical section
				go n.enterCriticalSection()
			}
		}
		// }
	}
}

// function to update the vector clock of a node
func (n *Node) updateVectorClock(msg Message) {
	for i := 0; i < numberOfNodes; i++ {
		if msg.senderVectorClock[i] > n.vectorClock[i] {
			n.vectorClock[i] = msg.senderVectorClock[i]
		}
	}
	n.vectorClock[n.id]++
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
	// if i do not want to enter the critical section, i will send a reply to the node that has requested to enter critical section
	myStatus := Message{
		senderID:          n.id,
		messageType:       "reply",
		senderVectorClock: n.vectorClock}
	for {
		if !n.isInCS {
			if !n.wantToEnterCS || (n.wantToEnterCS && ((vectorClockLessThan(&msg, &myStatus)) || n.checkIfReceivedAck(requester))) {
				requester.messageChan <- myStatus
				return
			}
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
	if len(n.priorityQueue) <= 1 {
		n.priorityQueue = make([]Message, 0)
	} else {
		n.priorityQueue = n.priorityQueue[1:]
	}
	n.wantToEnterCS = false
	n.isInCS = false
	fmt.Println("Node", n.id, "EXITING critical section...")
	for i := 0; i < numberOfNodes; i++ {
		if i != n.id {
			getNodeFromID(i).messageChan <- Message{
				senderID:          n.id,
				messageType:       "release",
				senderVectorClock: n.vectorClock,
			}
		}
	}
	n.time_exited = time.Now()
}

// function to automatically, periodically send a request to enter critical section to all the other nodes via the message channel
func (n *Node) requestToEnterCriticalSection() {
	// for {
	// time.Sleep(time.Duration(rand.Intn(20)) * time.Second)
	for !n.hasBeenInCS {
		n.time_requested = time.Now()
		if !n.wantToEnterCS {
			n.receivedReplies = make([]int, 0)
			// fmt.Println("Node", n.id, "requesting to enter critical section...")
			n.insertMessageIntoPQueue(Message{senderID: n.id,
				messageType:       "request",
				senderVectorClock: n.vectorClock})
			n.wantToEnterCS = true
			n.vectorClock[n.id]++
			VCsnapshot := n.vectorClock
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
func Q1P1() {
	fmt.Println("Q1P1 starting...")
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
			// fmt.Printf("testing\n")
			if (nodes[i].time_exited.Sub(nodes[i].time_requested)) > time.Duration(0) && timesInCS[i] == 0 {
				timesInCS[i] = nodes[i].time_exited.Sub(nodes[i].time_requested)
				// i want to print all times
				fmt.Printf("Node %d's time in CS: %v\n", i, nodes[i].time_exited.Sub(nodes[i].time_requested))
				fmt.Printf("Length of timesInCS: %d\n", len(timesInCS))
			}
		}
		// print length of timesInCS
	}
}
