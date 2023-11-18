package pset2part3

// // import (
// // 	"fmt"
// // 	"math/rand"
// // 	"time"
// // )

// // type Node struct {
// // 	id              int
// // 	messageChan     chan Message
// // 	vectorClock     [4]int
// // 	localQueue      []Message
// // 	isInCS          bool
// // 	wantToEnterCS   bool
// // 	canEnterCS      bool
// // 	receivedVoteIDs []int // slice of ids of received replies
// // 	haveVoted       bool
// // 	votedFor        int
// // }

// // type Message struct {
// // 	senderID          int
// // 	messageType       string // can be "request", "reply", "release"
// // 	senderVectorClock [4]int
// // }

// // func to update vector clock
// func updateVectorClock(localClock, receivedClock []int, nodeID int) {
// 	// Iterate over each element in the vector clock
// 	for i := range localClock {
// 		// Update each element by taking the maximum value from local and received clocks
// 		localClock[i] = max(localClock[i], receivedClock[i])
// 	}

// 	// Increment the element corresponding to the node making the update
// 	localClock[nodeID]++
// }

// // max function to find the maximum of two integers
// func max(a, b int) int {
// 	if a > b {
// 		return a
// 	}
// 	return b
// }

// func initNode(id int) Node {
// 	n := Node{}
// 	n.id = id
// 	n.messageChan = make(chan Message, 100)
// 	n.vectorClock[id] = 1
// 	n.localQueue = make([]Message, 0)
// 	n.receivedVoteIDs = make([]int, 0)
// 	n.haveVoted = false
// 	n.votedFor = -1
// 	n.canEnterCS = true
// 	return n
// }

// // func to get Node from id
// func getNodeFromID(id int) *Node {
// 	return &nodes[id]
// }

// func (n *Node) voteFor(requester *Node) {
// 	n.haveVoted = true
// 	n.votedFor = requester.id

// 	// update vector clock
// 	n.vectorClock[n.id]++

// 	requester.messageChan <- Message{
// 		senderID:          n.id,
// 		messageType:       "VOTE",
// 		senderVectorClock: n.vectorClock,
// 	}
// }

// func (n *Node) rescindVoteFor(requester *Node) {
// 	n.haveVoted = false
// 	n.votedFor = -1

// 	// update vector clock
// 	n.vectorClock[n.id]++

// 	VCsnapshot := n.vectorClock

// 	requester.messageChan <- Message{
// 		senderID:          n.id,
// 		messageType:       "RELEASE-VOTE",
// 		senderVectorClock: VCsnapshot,
// 	}
// }

// func vectorClockLessThan(smallerMessage, largerMessage *Message) bool {
// 	hasHigher := false
// 	hasLower := false
// 	for i := 0; i < numberOfNodes; i++ {
// 		if smallerMessage.senderVectorClock[i] > largerMessage.senderVectorClock[i] {
// 			hasHigher = true
// 		} else if smallerMessage.senderVectorClock[i] < largerMessage.senderVectorClock[i] {
// 			hasLower = true
// 		}
// 	}
// 	if hasHigher && hasLower {
// 		return smallerMessage.senderID < largerMessage.senderID
// 	} else if hasHigher && !hasLower {
// 		return false
// 	}
// 	return true
// }

// // enter critical section for a random amount of time
// func (n *Node) enterCriticalSection() {
// 	// print who voted for me
// 	// time.Sleep(time.Duration(1 * time.Second))

// 	fmt.Println("Node", n.id, "RECEIVED VOTES FROM", n.receivedVoteIDs)
// 	fmt.Println("Node", n.id, "entered critical section")
// 	n.isInCS = true

// 	// update vector clock
// 	n.vectorClock[n.id]++

// 	// we want to set everyone's canEnterCS to false
// 	for i := 0; i < numberOfNodes; i++ {
// 		getNodeFromID(i).canEnterCS = false
// 	}
// 	time.Sleep(time.Duration(rand.Intn(2)) * time.Second)
// 	// exit critical section
// 	n.wantToEnterCS = false
// 	n.isInCS = false
// 	// we want to set everyone's canEnterCS to true
// 	for i := 0; i < numberOfNodes; i++ {
// 		getNodeFromID(i).canEnterCS = true
// 	}
// 	// set hasVoted to false
// 	n.haveVoted = false
// 	// set votedFor to -1
// 	n.votedFor = -1

// 	// update receivedVoteIDs
// 	n.receivedVoteIDs = make([]int, 0)

// 	fmt.Println("Node", n.id, "EXITING critical section...")

// 	// send release messages to all nodes
// 	for i := 0; i < numberOfNodes; i++ {
// 		// if i != n.id {
// 		getNodeFromID(i).messageChan <- Message{
// 			senderID:          n.id,
// 			messageType:       "RELEASE-VOTE",
// 			senderVectorClock: n.vectorClock,
// 		}
// 		// }
// 	}
// }

// func (n *Node) listenForMessages() {
// 	for {
// 		msg := <-n.messageChan

// 		// right after receiving a message, update my vector clock
// 		updateVectorClock(n.vectorClock[:], msg.senderVectorClock[:], n.id)

// 		if msg.messageType == "REQUEST" {
// 			// I have just received a vote request
// 			// fmt.Println("Node", n.id, "received vote REQUEST from Node", msg.senderID)
// 			// if I have not voted yet, vote for requester
// 			if !n.haveVoted {
// 				n.voteFor(getNodeFromID(msg.senderID))
// 			} else if n.haveVoted && vectorClockLessThan(&msg, &Message{senderID: n.id, senderVectorClock: n.vectorClock}) { // I have already voted but this vote request is timestamped earlier than my vote, so I rescind my vote and vote for requester
// 				n.rescindVoteFor(getNodeFromID(n.votedFor))
// 				// I have already voted, so I just need to add this incoming vote request to my local queue
// 			} else {
// 				n.localQueue = append(n.localQueue, msg)
// 			}

// 		} else if msg.messageType == "VOTE" {
// 			fmt.Println("Node", n.id, "received VOTE from Node", msg.senderID)

// 			// received a vote but i've already exited the critical section, so i want to return the vote
// 			if !n.wantToEnterCS && !n.isInCS {
// 				n.rescindVoteFor(getNodeFromID(msg.senderID))
// 			}

// 			n.receivedVoteIDs = append(n.receivedVoteIDs, msg.senderID)
// 			if len(n.receivedVoteIDs) > numberOfNodes/2 && !n.isInCS && n.canEnterCS {
// 				// I have received a majority of votes, so I can enter the critical section
// 				n.isInCS = true
// 				n.wantToEnterCS = false
// 				// enter critical section and then exit
// 				n.enterCriticalSection()
// 			}
// 		} else if msg.messageType == "RELEASE-VOTE" {
// 			// I have received a release vote message, so I can now vote for the next node in my local queue
// 			// fmt.Println("Node", n.id, "received RELEASE-VOTE from Node", msg.senderID)
// 			// but i first need to pop the requester from my local queue
// 			// if len(n.localQueue) > 0 {
// 			// 	n.localQueue = n.localQueue[1:]
// 			// }

// 			for i, m := range n.localQueue {
// 				if m.senderID == msg.senderID {
// 					n.localQueue = append(n.localQueue[:i], n.localQueue[i+1:]...)
// 				}
// 			}
// 			// i also need to remove the requester from my receivedVoteIDs
// 			for i, id := range n.receivedVoteIDs {
// 				if id == msg.senderID {
// 					n.receivedVoteIDs = append(n.receivedVoteIDs[:i], n.receivedVoteIDs[i+1:]...)
// 				}
// 			}
// 			// i need to change my hasVoted status
// 			n.haveVoted = false

// 			// update vector clock
// 			updateVectorClock(n.vectorClock[:], msg.senderVectorClock[:], n.id)

// 			if len(n.localQueue) > 0 {
// 				n.voteFor(getNodeFromID(n.localQueue[0].senderID))
// 				n.localQueue = n.localQueue[1:]
// 			} else {
// 				// I have no more vote requests in my local queue, so I can rescind my vote
// 				n.haveVoted = false
// 				n.votedFor = -1
// 			}
// 		}
// 	}
// }

// func (n *Node) requestToEnterCriticalSection() {
// 	for {
// 		time.Sleep(time.Duration(rand.Intn(4)) * time.Second)
// 		if !n.wantToEnterCS && !n.isInCS && n.canEnterCS {
// 			// reset receivedVoteIDs
// 			n.receivedVoteIDs = make([]int, 0)
// 			// add self to receivedVoteIDs
// 			n.receivedVoteIDs = append(n.receivedVoteIDs, n.id)
// 			// mark self as having voted
// 			n.haveVoted = true
// 			// vote for self
// 			n.votedFor = n.id
// 			fmt.Println("Node", n.id, "REQUESTING to enter critical section...")
// 			n.wantToEnterCS = true
// 			// update vector clock
// 			n.vectorClock[n.id]++

// 			VCsnapshot := n.vectorClock
// 			for i := 0; i < numberOfNodes; i++ {
// 				if i != n.id {
// 					getNodeFromID(i).messageChan <- Message{
// 						senderID:          n.id,
// 						messageType:       "REQUEST",
// 						senderVectorClock: VCsnapshot,
// 					}
// 				}
// 			}
// 		}
// 	}
// }

// var nodes []Node
// var numberOfNodes = 4

// func Q1p3() {
// 	fmt.Println("Q1p3 starting...")
// 	for i := 0; i < numberOfNodes; i++ {
// 		nodes = append(nodes, initNode(i))
// 	}
// 	for i := 0; i < numberOfNodes; i++ {
// 		go nodes[i].listenForMessages()
// 		go nodes[i].requestToEnterCriticalSection()
// 	}
// 	// wait for goroutines to finish
// 	for {
// 		time.Sleep(1 * time.Second)
// 	}
// }
