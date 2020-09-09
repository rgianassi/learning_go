package main

import "testing"

func TestHello(t *testing.T) {
	x := 1
	y := 2

	sum := x + y

	if sum != x+y {
		t.Fatal("sum must be a sum", x, y, sum)
	}
}
