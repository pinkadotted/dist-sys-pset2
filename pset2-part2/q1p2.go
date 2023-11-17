package pset2part2

import (
	"fmt"
	"math/rand"
	"time"
)

type Node struct {
	id              int
	messageChan     chan Message
	vectorClock     [5]int
	priorityQueue   []Message
	isInCS          bool
	wantToEnterCS   bool
	receivedReplies []int // slice of ids of received replies
}

type Message struct {
	senderID          int
	messageType       string // can be "request", "reply", "release"
	senderVectorClock [5]int
}

func initNode(id int) Node {
	n := Node{}
	n.id = id
	n.messageChan = make(chan Message, 100)
	n.vectorClock[id] = 0
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
// func (n *Node) checkIfReceivedAck(requester *Node) bool {
// 	for _, reply := range n.receivedReplies {
// 		if reply == requester.id {
// 			return true
// 		}
// 	}
// 	return false
// }

// MAIN FUNCTIONS:
// function to listen for messages
func (n *Node) listenForMessages() {
	for {
		msg := <-n.messageChan
		// fmt.Printf("%d: received message from %d of type %s\n", n.id, msg.senderID, msg.messageType)
		// n.updateVectorClock(msg)
		if msg.messageType == "request" {
			go n.replyToRequest(msg)
			// } else if msg.messageType == "release" {
			// 	if len(n.priorityQueue) == 1 {
			// 		n.priorityQueue = make([]Message, 0)
			// 	} else {
			// 		n.priorityQueue = n.priorityQueue[1:]
			// 		fmt.Println("Node", n.id, "has new priority queue after RELEASE MSG:", n.priorityQueue)
			// 	}
			// 	if n.wantToEnterCS && len(n.receivedReplies) == numberOfNodes-1 && len(n.priorityQueue) > 0 && n.priorityQueue[0].senderID == n.id {
			// 		// enter critical section
			// 		go n.enterCriticalSection()
			// 	}
		} else if msg.messageType == "reply" {
			n.receivedReplies = append(n.receivedReplies, msg.senderID)
			// fmt.Printf("%d RECEIVED REPLIES: %v\n", n.id, n.receivedReplies)
			if len(n.receivedReplies) == numberOfNodes-1 {
				// if len(n.receivedReplies) == numberOfNodes-1 && len(n.priorityQueue) > 0 && n.priorityQueue[0].senderID == n.id {
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
	// n.vectorClock[n.id]++
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
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	n.vectorClock[n.id]++
	n.receivedReplies = make([]int, 0)
	// if len(n.priorityQueue) <= 1 {
	// 	n.priorityQueue = make([]Message, 0)
	// } else {
	// 	n.priorityQueue = n.priorityQueue[1:]
	// }
	n.wantToEnterCS = false
	n.isInCS = false
	fmt.Println("Node", n.id, "EXITING critical section...")
	// for i := 0; i < numberOfNodes; i++ {
	// 	if i != n.id {
	// 		getNodeFromID(i).messageChan <- Message{
	// 			senderID:          n.id,
	// 			messageType:       "release",
	// 			senderVectorClock: n.vectorClock,
	// 		}
	// 	}
	// }
	for _, message := range n.priorityQueue {
		n.replyToRequest(message)
	}
	n.priorityQueue = make([]Message, 0)
}

// function to automatically, periodically send a request to enter critical section to all the other nodes via the message channel
func (n *Node) requestToEnterCriticalSection() {
	for {
		time.Sleep(time.Duration(rand.Intn(20)) * time.Second)
		if !n.wantToEnterCS {
			n.receivedReplies = make([]int, 0)
			// n.insertMessageIntoPQueue(Message{senderID: n.id,
			// 	messageType:       "request",
			// 	senderVectorClock: n.vectorClock})
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
	}
}

var nodes []Node
var numberOfNodes = 5

// function to initialize all the nodes, and start the listening for messages and request to enter critical section goroutines, and randomly send requests to enter critical section
func Q1p2() {
	fmt.Println("Q1p2 starting...")
	for i := 0; i < numberOfNodes; i++ {
		nodes = append(nodes, initNode(i))
	}
	for i := 0; i < numberOfNodes; i++ {
		go nodes[i].listenForMessages()
		go nodes[i].requestToEnterCriticalSection()
	}
	// wait for goroutines to finish
	for {
		time.Sleep(1 * time.Second)
	}
}
