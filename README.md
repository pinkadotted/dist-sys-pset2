## Distributed Systems Problem Set 2

This repository contains implementations of various distributed mutual exclusion protocols using the Go programming language. The implemented protocols include Lamport's shared priority queue both with and without Ricart and Agrawala's optimization, as well as a Voting Protocol with deadlock avoidance.

### 1. Lamport's shared priority queue without Ricart Agrawala's optimisation

To run this part, go to main.go in the root dir, uncomment out the part that is between "START OF PART 1" and "END OF PART 1", and run 
```go run main.go``` 

### 2. Lamport's shared priority queue with Ricart Agrawala's optimisation

To run this part, go to main.go in the root dir, uncomment out the part that is between "START OF PART 2" and "END OF PART 2", and run 
```go run main.go``` 

### 3. Voting Protocol with deadlock avoidance 

To run this part, go to main.go in the root dir, uncomment out the part that is between "START OF PART 3" and "END OF PART 4", and run 
```go run main.go``` 

### Comparing performances

For this part, just uncomment everything and run it. You should see 30 entries in the terminal, 1 for each node for each protocol. 

The final timings for each protocol is shown all the way at the bottom. 

### Experiment Setup
- Number of Nodes: 10
- Protocols:
    - Lamport’s Shared Priority Queue without Ricart and Agrawala’s Optimization
    - Lamport’s Shared Priority Queue with Ricart and Agrawala’s Optimization
    - Voting Protocol with Deadlock Avoidance

### Conclusion

Overall, I think that the protocol implemented in part 2 is quicker than that in part 1, and part 3 is quicker than part 2. So part 1 is the slowest, followed by part 2, followed by part 3. 

This problem set was done in collaboration with Cheong Cher Lynn. 