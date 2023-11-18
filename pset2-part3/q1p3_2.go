package pset2part3

import (
	"fmt"
	"math/rand"
	"time"
)

type Node struct {
	id            int
	messageChan   chan Message
	vectorClock   [4]int
	localQueue    []Message
	isInCS        bool
	wantToEnterCS bool
	// canEnterCS      bool
	receivedVoteIDs []int // slice of ids of received replies
	hasVoted        bool
	votedFor        Message
}

type Message struct {
	senderID          int
	messageType       string // can be "REQUEST-VOTE", "GIVE-VOTE", "RESCIND-VOTE", "RELEASE-VOTE"
	senderVectorClock [4]int
}

// func to initialize a node
// TODO: CHECK THIS FUNCTION

func initNode(id int) Node {
	n := Node{}
	n.id = id
	n.isInCS = false
	n.wantToEnterCS = false
	n.messageChan = make(chan Message, 100)
	n.vectorClock[id] = 0
	n.localQueue = make([]Message, 0)
	n.receivedVoteIDs = make([]int, 0)
	n.hasVoted = false
	n.votedFor = Message{}
	return n
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

// function to insert message into queue

func (n *Node) insertMessageIntoQueue(m Message) {
	// fmt.Println("Node", n.id, "inserting message into priority queue...")
	for i, msg := range n.localQueue {
		if vectorClockLessThan(&m, &msg) {
			n.localQueue = append(n.localQueue[:i], append([]Message{m}, n.localQueue[i:]...)...)
			// fmt.Printf("Node %d has new priority queue: %v\n", n.id, n.priorityQueue)
			return
		}
	}
	n.localQueue = append(n.localQueue, m)
	// fmt.Printf("Node %d has new priority queue: %v\n", n.id, n.priorityQueue)
}

// function where a node wants to enter CS and sends a request

func (n *Node) sendRequest() {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	n.wantToEnterCS = true
	n.vectorClock[n.id]++
	vcSnapshot := n.vectorClock
	requestMessage := Message{
		senderID:          n.id,
		senderVectorClock: vcSnapshot,
		messageType:       "REQUEST-VOTE",
	}
	for _, node := range nodes {
		if node.id != n.id {
			node.messageChan <- requestMessage
		}
	}
	n.votedFor = requestMessage
	n.hasVoted = true
	n.receivedVoteIDs = append(n.receivedVoteIDs, n.id)
}

// function to enter CS
func (n *Node) enterCS() {
	n.isInCS = true

	fmt.Println("Node", n.id, "ENTERING critical section...")
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

	// exit critical section
	n.isInCS = false
	n.wantToEnterCS = false
	n.hasVoted = false
	fmt.Println("Node", n.id, "EXITING critical section...")
	// send release messages to all nodes
	for _, node := range nodes {
		node.messageChan <- Message{
			senderID:          n.id,
			senderVectorClock: n.vectorClock,
			messageType:       "RELEASE-VOTE",
		}
	}

}

// function to receive a request
// TODO: CHECK THIS FUNCTION

func (n *Node) receiveRequest(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := n.vectorClock

	if !n.hasVoted {
		n.votedFor = msg
		n.hasVoted = true
		msg.senderVectorClock = vcSnapshot
		n.giveVote(msg)

	} else { // I have already voted, so I need to either rescind my vote or add the request to my local queue
		// if incoming message has a smaller vector clock than my vote, I rescind my vote and vote for requester
		if vectorClockLessThan(&msg, &n.votedFor) {
			n.votedFor.senderVectorClock = vcSnapshot
			n.sendRescindVote(n.votedFor)
			// add the rescindee back to my local queue
			n.insertMessageIntoQueue(n.votedFor)
		}
		// add this new request to the correct position of my local queue
		n.insertMessageIntoQueue(msg)
	}
}

// function to give vote
// TODO: CHECK THIS FUNCTION

func (n *Node) giveVote(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := n.vectorClock
	n.votedFor = msg
	n.hasVoted = true
	giveVoteMessage := Message{
		senderID:          n.id,
		senderVectorClock: vcSnapshot,
		messageType:       "GIVE-VOTE",
	}
	nodes[msg.senderID].messageChan <- giveVoteMessage
}

// function to receive vote
// TODO: CHECK THIS FUNCTION

func (n *Node) receiveVote(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := n.vectorClock
	n.receivedVoteIDs = append(n.receivedVoteIDs, msg.senderID)
	// i want to check if I want to enter CS and I am not already in CS
	if n.wantToEnterCS && !n.isInCS {
		// Check if I have received a majority of votes
		if len(n.receivedVoteIDs) > numberOfNodes/2 {
			// I have received a majority of votes, so I can enter the critical section
			n.isInCS = true
			n.wantToEnterCS = false
			// enter critical section and then exit
			n.enterCS()
		}
	} else {
		msg.senderVectorClock = vcSnapshot
		n.sendReleaseVote(msg)
	}
}

// function where requester sends a release vote to the sender of this message (a voter)
// TODO: CHECK THIS FUNCTION

func (n *Node) sendReleaseVote(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := n.vectorClock
	// remove sender's vote from my received votes
	for i, vote := range n.receivedVoteIDs {
		if vote == msg.senderID {
			if len(n.receivedVoteIDs) > 1 {
				n.receivedVoteIDs = append(n.receivedVoteIDs[:i], n.receivedVoteIDs[i+1:]...)
			} else {
				n.receivedVoteIDs = make([]int, 0)
			}
		}
	}
	releaseVoteMessage := Message{
		senderID:          n.id,
		senderVectorClock: vcSnapshot,
		messageType:       "RELEASE-VOTE",
	}
	nodes[msg.senderID].messageChan <- releaseVoteMessage
}

// function where voter receives a released vote from the node it voted for
// TODO: CHECK THIS FUNCTION

func (n *Node) receiveReleaseVote(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := n.vectorClock
	n.hasVoted = false
	n.votedFor = Message{}
	// then I check my local queue to see if there is a request that I can vote for
	if len(n.localQueue) > 0 {
		if !n.hasVoted {
			n.votedFor = n.localQueue[0]
			n.hasVoted = true
			n.localQueue[0].senderVectorClock = vcSnapshot
			n.giveVote(n.localQueue[0])
			n.localQueue = n.localQueue[1:]
		}
	} // else do nothing
}

// function to send rescind vote
// TODO: CHECK THIS FUNCTION

func (n *Node) sendRescindVote(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := n.vectorClock
	rescindVoteMessage := Message{
		senderID:          n.id,
		senderVectorClock: vcSnapshot,
		messageType:       "RESCIND-VOTE",
	}
	// send rescind vote to the node I voted for
	nodes[msg.senderID].messageChan <- rescindVoteMessage

}

// function to receive rescind vote
// TODO: CHECK THIS FUNCTION

func (n *Node) receiveRescindVote(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := n.vectorClock
	// if I'm not in CS, remove rescinder's vote from my received votes
	if !n.isInCS {
		// for i, vote := range n.receivedVoteIDs {
		// 	if vote == msg.senderID {
		// 		if len(n.receivedVoteIDs) > 1 {
		// 			n.receivedVoteIDs = append(n.receivedVoteIDs[:i], n.receivedVoteIDs[i+1:]...)
		// 		} else {
		// 			n.receivedVoteIDs = make([]int, 0)
		// 		}
		// 	}
		// }
		msg.senderVectorClock = vcSnapshot
		n.sendReleaseVote(msg)
	}
}

// function to listen for messages
// TODO: CHECK THIS FUNCTION

func (n *Node) listenForMessages() {
	for {
		msg := <-n.messageChan
		// n.updateVectorClock(msg)
		switch msg.messageType {
		case "REQUEST-VOTE":
			n.receiveRequest(msg)
		case "GIVE-VOTE":
			n.receiveVote(msg)
		case "RESCIND-VOTE":
			n.receiveRescindVote(msg)
		case "RELEASE-VOTE":
			n.receiveReleaseVote(msg)
		}
	}
}

var nodes []Node
var numberOfNodes = 4

// main function
// TODO: CHECK THIS FUNCTION

func Q1P3_2() {
	nodes = make([]Node, numberOfNodes)
	for i := 0; i < numberOfNodes; i++ {
		nodes[i] = initNode(i)
	}
	for i := 0; i < numberOfNodes; i++ {
		go nodes[i].listenForMessages()
		go nodes[i].sendRequest()
	}
	time.Sleep(100 * time.Second)
}
