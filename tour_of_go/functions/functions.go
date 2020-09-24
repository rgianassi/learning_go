package main

import (
	"fmt"
	"math"
)

func compute(fn func(float64, float64) float64) float64 {
	return fn(3, 4)
}

func adder() func(int) int {
	sum := 0
	return func(x int) int {
		sum += x
		return sum
	}
}

func functionValues() {
	hypot := func(x, y float64) float64 {
		return math.Sqrt(x*x + y*y)
	}
	fmt.Println(hypot(5, 12))

	fmt.Println(compute(hypot))
	fmt.Println(compute(math.Pow))
}

func functionClosures() {
	pos, neg := adder(), adder()
	for i := 0; i < 10; i++ {
		fmt.Println(
			pos(i),
			neg(-2*i),
		)
	}
}

// fibonacci is a function that returns
// a function that returns an int.
func fibonacci() func() int {
	call_counter := 0
	last_value := 1
	last_to_one_value := 0
	return func() int {
		switch {
		case call_counter == 0:
			call_counter += 1
			return 0
		case call_counter == 1:
			call_counter += 1
			return 1
		default:
			result := last_to_one_value + last_value
			last_to_one_value = last_value
			last_value = result
			return result
		}
	}
}

func fibonacciClosure() {
	f := fibonacci()
	for i := 0; i < 10; i++ {
		fmt.Println(f())
	}
}

func main() {
	functionValues()
	functionClosures()
	fibonacciClosure()
}
