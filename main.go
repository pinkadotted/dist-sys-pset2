package main

import (
	"fmt"
	pset2part1 "pset2/pset2-part1"
	"time"
)

func main() {
	// START OF PART 1
	start_part1 := time.Now()
	pset2part1.Q1P1()
	end_part1 := time.Now()
	fmt.Printf("Part 1 took %v\n", end_part1.Sub(start_part1))
	// END OF PART 1

	// START OF PART 2
	// start_part2 := time.Now()
	// pset2part2.Q1P2()
	// end_part2 := time.Now()
	// fmt.Printf("Part 2 took %v\n", end_part2.Sub(start_part2))
	// END OF PART 2

	// START OF PART 3
	// start_part3 := time.Now()
	// pset2part3.Q1P3()
	// end_part3 := time.Now()
	// fmt.Printf("Part 3 took %v\n", end_part3.Sub(start_part3))
	// END OF PART 3
}
