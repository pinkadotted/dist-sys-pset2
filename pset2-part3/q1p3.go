package pset2part3

import (
	"fmt"
	"time"
)

type Node struct {
	id            int
	messageChan   chan Message
	vectorClock   [10]int
	localQueue    []Message
	isInCS        bool
	wantToEnterCS bool
	// canEnterCS      bool
	receivedVoteIDs []int // slice of ids of received replies
	hasVoted        bool
	votedFor        Message
	time_requested  time.Time
	time_exited     time.Time
	hasBeenInCS     bool
}

type Message struct {
	senderID          int
	messageType       string // can be "REQUEST-VOTE", "GIVE-VOTE", "RESCIND-VOTE", "RELEASE-VOTE"
	senderVectorClock []int
}

// func to initialize a node
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
	for !n.hasBeenInCS {
		// time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
		n.time_requested = time.Now()

		if !n.wantToEnterCS && !n.isInCS {
			n.wantToEnterCS = true
			n.vectorClock[n.id]++
			vcSnapshot := append([]int(nil), n.vectorClock[:]...)
			requestMessage := Message{
				senderID:          n.id,
				senderVectorClock: vcSnapshot,
				messageType:       "REQUEST-VOTE",
			}
			if !n.hasVoted {
				n.votedFor = requestMessage
				copy(n.votedFor.senderVectorClock, requestMessage.senderVectorClock)
				n.hasVoted = true
				n.receivedVoteIDs = append(n.receivedVoteIDs, n.id)
			} else {
				n.insertMessageIntoQueue(requestMessage)
			}
			fmt.Printf("------------------  Node %d REQUEST-VOTE: %v\n", n.id, n.vectorClock)
			for _, node := range nodes {
				if node.id != n.id {
					node.messageChan <- requestMessage
				}
			}
		}
		n.hasBeenInCS = true
	}
}

// function to enter CS
func (n *Node) enterCS() {
	n.isInCS = true

	fmt.Println("<<<<<<<<<<<<<<<<<<< Node", n.id, "ENTERING critical section...")
	time.Sleep(time.Duration(1) * time.Second)

	// exit critical section
	fmt.Println(">>>>>>>>>>>>>>>>>>> Node", n.id, "EXITING critical section...")

	// send release messages to all nodes
	// fmt.Printf("Node %d releasing votes: %v\n", n.id, n.receivedVoteIDs)
	for _, nodeid := range n.receivedVoteIDs {
		nodes[nodeid].messageChan <- Message{
			senderID:          n.id,
			senderVectorClock: n.vectorClock[:],
			messageType:       "RELEASE-VOTE",
		}
	}
	n.receivedVoteIDs = make([]int, 0)
	n.isInCS = false
	n.wantToEnterCS = false
	// n.hasBeenInCS = true
	n.time_exited = time.Now()
}

// function to receive a request
func (n *Node) receiveRequest(msg Message) {
	n.updateVectorClock(msg)
	// n.vectorClock[n.id]--
	vcSnapshot := n.vectorClock[:]

	if !n.hasVoted {
		// n.votedFor = msg
		// n.hasVoted = true
		msg.senderVectorClock = vcSnapshot
		n.giveVote(msg)

	} else { // I have already voted, so I need to either rescind my vote or add the request to my local queue
		// if incoming message has a smaller vector clock than my vote, I rescind my vote and vote for requester
		if vectorClockLessThan(&msg, &n.votedFor) {
			// fmt.Printf("INCOMING: %v, CURRENT: %v\n", msg.senderVectorClock, n.votedFor.senderVectorClock)
			n.votedFor.senderVectorClock = vcSnapshot
			n.sendRescindVote(n.votedFor)

			// add the rescindee back to my local queue
			// n.insertMessageIntoQueue(n.votedFor) //// BUG: should only insert if rescind successful.
			// } else {
			// 	// add this new request to the correct position of my local queue
			// 	n.insertMessageIntoQueue(msg)
		}
		n.insertMessageIntoQueue(msg)
		// fmt.Printf("Node %d's updated QUEUE: %v\n", n.id, n.localQueue)
	}
}

// function to give vote
func (n *Node) giveVote(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := append([]int(nil), n.vectorClock[:]...)
	n.votedFor = msg
	n.votedFor.senderVectorClock = append(make([]int, 0), msg.senderVectorClock...)
	n.hasVoted = true
	giveVoteMessage := Message{
		senderID:          n.id,
		senderVectorClock: vcSnapshot,
		messageType:       "GIVE-VOTE",
	}
	nodes[msg.senderID].messageChan <- giveVoteMessage
	// fmt.Printf("Node %d GIVE-VOTE to: %v\n", n.id, n.votedFor.senderID)
}

// function to receive vote
func (n *Node) receiveVote(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := append([]int(nil), n.vectorClock[:]...)
	n.receivedVoteIDs = append(n.receivedVoteIDs, msg.senderID)
	// fmt.Printf("Node %d's RECEIVED-VOTES: %v\n", n.id, n.receivedVoteIDs)
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
	} else if !n.wantToEnterCS {
		msg.senderVectorClock = vcSnapshot
		n.sendReleaseVote(msg)
		// fmt.Printf("Node %d doesn't need Node %d's vote anymore.\n", n.id, msg.senderID)
	}
}

// function where requester sends a release vote to the sender of this message (a voter)
func (n *Node) sendReleaseVote(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := append([]int(nil), n.vectorClock[:]...)
	// remove sender's vote from my received votes
	var updatedSlice []int
	for _, vote := range n.receivedVoteIDs {
		if vote != msg.senderID {
			// if len(n.receivedVoteIDs) > 1 {
			// 	n.receivedVoteIDs = append(n.receivedVoteIDs[:i], n.receivedVoteIDs[i+1:]...)
			// } else {
			// 	n.receivedVoteIDs = make([]int, 0)
			// }
			updatedSlice = append(updatedSlice, vote)
		} else {
			releaseVoteMessage := Message{
				senderID:          n.id,
				senderVectorClock: vcSnapshot,
				messageType:       "RELEASE-VOTE",
			}
			nodes[msg.senderID].messageChan <- releaseVoteMessage
			// fmt.Printf("Node %d RELEASE-VOTE back to: %v\n", n.id, msg.senderID)
		}
	}
	n.receivedVoteIDs = updatedSlice
}

// function where voter receives a released vote from the node it voted for
func (n *Node) receiveReleaseVote(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := append([]int(nil), n.vectorClock[:]...)
	n.hasVoted = false
	n.votedFor = Message{}
	// fmt.Printf("Node %d CAN VOTE AGAIN!\n", n.id)
	if len(n.localQueue) > 0 {
		if !n.hasVoted {
			n.votedFor = n.localQueue[0]
			n.votedFor.senderVectorClock = append(make([]int, 0), msg.senderVectorClock...)
			n.hasVoted = true
			n.localQueue[0].senderVectorClock = vcSnapshot
			n.giveVote(n.localQueue[0])
			n.localQueue = n.localQueue[1:]
			// fmt.Printf("Node %d's local queue: %v\n", n.id, n.localQueue)
		}
	} // else do nothing
}

// function to send rescind vote
func (n *Node) sendRescindVote(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := append([]int(nil), n.vectorClock[:]...)
	rescindVoteMessage := Message{
		senderID:          n.id,
		senderVectorClock: vcSnapshot,
		messageType:       "RESCIND-VOTE",
	}
	// send rescind vote to the node I voted for
	nodes[msg.senderID].messageChan <- rescindVoteMessage
	// fmt.Printf("Node %d wants to RESCIND-VOTE from: %d\n", n.id, msg.senderID)
}

// function to receive rescind vote
func (n *Node) receiveRescindVote(msg Message) {
	n.updateVectorClock(msg)
	vcSnapshot := append([]int(nil), n.vectorClock[:]...)
	// if I'm not in CS, remove rescinder's vote from my received votes
	// fmt.Printf("Node %d received RESCIND-VOTE from: %v\n", n.id, msg.senderID)
	if !n.isInCS {
		msg.senderVectorClock = vcSnapshot
		n.sendReleaseVote(msg)
	}
}

// function to listen for messages
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
var numberOfNodes = 10

// make a map of node ids to their times in CS
var timesInCS = make(map[int]time.Duration)

// main function
func Q1P3() {
	fmt.Println("Running Q1P3...")
	nodes = make([]Node, numberOfNodes)
	for i := 0; i < numberOfNodes; i++ {
		nodes[i] = initNode(i)
	}
	for i := 0; i < numberOfNodes; i++ {
		go nodes[i].listenForMessages()
		go nodes[i].sendRequest()
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
