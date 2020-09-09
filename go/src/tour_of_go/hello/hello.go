package main

import (
	"fmt"
	"math/rand"
)

func swap(x, y string) (string, string) {
	return y, x
}

func main() {
	fmt.Println("My favorite number is", rand.Intn(10))
	fmt.Println("Hello, 世界")

	a, b := swap("hello", "world")
	fmt.Println(a, b)
}
