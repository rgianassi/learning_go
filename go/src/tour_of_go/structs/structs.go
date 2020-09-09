package main

import (
	"fmt"
	"strings"

	"golang.org/x/tour/pic"
)

type Vertex struct {
	X int
	Y int
}

var (
	v1 = Vertex{1, 2}  // has type Vertex
	v2 = Vertex{X: 1}  // Y:0 is implicit
	v3 = Vertex{}      // X:0 and Y:0
	p  = &Vertex{1, 2} // has type *Vertex
)

func pointers() {
	fmt.Println("pointers")

	i, j := 42, 2701

	p := &i         // point to i
	fmt.Println(*p) // read i through the pointer
	*p = 21         // set i through the pointer
	fmt.Println(i)  // see the new value of i

	p = &j         // point to j
	*p = *p / 37   // divide j through the pointer
	fmt.Println(j) // see the new value of j
}

func structs() {
	fmt.Println("structs")

	v := Vertex{1, 2}
	fmt.Println(v)
	v.X = 4
	fmt.Println(v)
	pp := &v
	pp.X = 1e9
	fmt.Println(v)

	fmt.Println(v1, p, v2, v3)
}

func sliceLengthCapacity() {
	fmt.Println("sliceLengthCapacity")

	s := []int{2, 3, 5, 7, 11, 13}
	printSlice(s)

	// Slice the slice to give it zero length.
	s = s[:0]
	printSlice(s)

	// Extend its length.
	s = s[:4]
	printSlice(s)

	// Drop its first two values.
	s = s[2:]
	printSlice(s)

	s2 := s[2:4]
	printSlice(s2)

	// The following panics because out of capacity
	// s2 = s2[2:4]
}

func printSlice(s []int) {
	fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
}

func makeSlices() {
	fmt.Println("makeSlices")

	a := make([]int, 5)
	printSlice(a)

	b := make([]int, 0, 5)
	printSlice(b)

	c := b[:2]
	printSlice(c)

	d := c[2:5]
	printSlice(d)
}

func arrays() {
	var a [2]string
	a[0] = "Hello"
	a[1] = "World"
	fmt.Println(a[0], a[1])
	fmt.Println(a)

	primes := [6]int{2, 3, 5, 7, 11, 13}
	fmt.Println(primes)

	var s []int = primes[1:4]
	fmt.Println(s)

	s = s[:2] // reassigning
	fmt.Println(s)

	s = s[1:] // reassigning
	fmt.Println(s)

	names := [4]string{
		"John",
		"Paul",
		"George",
		"Ringo",
	}
	fmt.Println(names)

	aa := names[0:2]
	bb := names[1:3]
	fmt.Println(aa, bb)

	bb[0] = "XXX"
	fmt.Println(aa, bb)
	fmt.Println(names)

	q := []int{2, 3, 5, 7, 11, 13}
	fmt.Println(q)

	r := []bool{true, false, true, true, false, true}
	fmt.Println(r)

	ss := []struct {
		i int
		b bool
	}{
		{2, true},
		{3, false},
		{5, true},
		{7, true},
		{11, false},
		{13, true},
	}
	fmt.Println(ss)
}

func slicesOfSlices() {
	fmt.Println("slicesOfSlices")

	// Create a tic-tac-toe board.
	board := [][]string{
		[]string{"_", "_", "_"},
		[]string{"_", "_", "_"},
		[]string{"_", "_", "_"},
	}

	// The players take turns.
	board[0][0] = "X"
	board[2][2] = "O"
	board[1][2] = "X"
	board[1][0] = "O"
	board[0][2] = "X"

	for i := 0; i < len(board); i++ {
		fmt.Printf("%s\n", strings.Join(board[i], " "))
	}
}

func slicesAppend() {
	fmt.Println("slicesAppend")

	var s []int
	printSlice(s)

	// append works on nil slices.
	s = append(s, 0)
	printSlice(s)

	// The slice grows as needed.
	s = append(s, 1)
	printSlice(s)

	// We can add more than one element at a time.
	s = append(s, 2, 3, 4)
	printSlice(s)
}

var pow = []int{1, 2, 4, 8, 16, 32, 64, 128}

func ranges() {
	for i, v := range pow {
		fmt.Printf("2**%d = %d\n", i, v)
	}
	pow := make([]int, 10)
	for i := range pow {
		pow[i] = 1 << uint(i) // == 2**i
	}
	for _, value := range pow {
		fmt.Printf("%d\n", value)
	}
}

func Pic(dx, dy int) [][]uint8 {
	var image [][]uint8
	for y := 0; y < dy; y++ {
		var imageLine []uint8
		for x := 0; x < dx; x++ {
			//var v = uint8((x + y) / 2)
			//var v = uint8(x * y)
			//var v = uint8(x ^ y)
			var v = uint8(x - y)
			imageLine = append(imageLine, v)
		}
		image = append(image, imageLine)
	}
	return image
}

func main() {
	pointers()
	structs()
	arrays()
	sliceLengthCapacity()
	makeSlices()
	slicesOfSlices()
	slicesAppend()
	ranges()
	pic.Show(Pic)
}
