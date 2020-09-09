package main

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"golang.org/x/tour/pic"
	"golang.org/x/tour/reader"
)

type Vertex struct {
	X, Y float64
}

type MyFloat float64

type Abser interface {
	Abs() float64
}

type I interface {
	M()
}

type T struct {
	S string
}

// This method means type T implements the interface I,
// but we don't need to explicitly declare that it does so.
func (t *T) M() {
	fmt.Println(t.S)
}

type F float64

func (f F) M() {
	fmt.Println(f)
}

func describe(i I) {
	fmt.Printf("(%v, %T)\n", i, i)
}

func (f MyFloat) Abs() float64 {
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}

func (v *Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v *Vertex) Scale(f float64) {
	v.X = v.X * f
	v.Y = v.Y * f
}

func ScaleFunc(v *Vertex, f float64) {
	v.X = v.X * f
	v.Y = v.Y * f
}

func methods() {
	v := Vertex{3, 4}
	fmt.Println(v.Abs())
}

func methods2() {
	f := MyFloat(-math.Sqrt2)
	fmt.Println(f.Abs())
}

func pointerReceivers() {
	v := Vertex{3, 4}
	v.Scale(10)
	fmt.Println(v.Abs())
}

func pointerIndirection() {
	v := Vertex{3, 4}
	v.Scale(2)
	ScaleFunc(&v, 10)

	p := &Vertex{4, 3}
	p.Scale(3)
	ScaleFunc(p, 8)

	fmt.Println(v, p)
}

func interfaces() {
	var a Abser
	f := MyFloat(-math.Sqrt2)
	v := Vertex{3, 4}

	a = f  // a MyFloat implements Abser
	a = &v // a *Vertex implements Abser

	// In the following line, v is a Vertex (not *Vertex)
	// and does NOT implement Abser.
	//a = v

	fmt.Println(a.Abs())
}

func interfaceValues() {
	var i I

	i = &T{"Hello"}
	describe(i)
	i.M()

	i = F(math.Pi)
	describe(i)
	i.M()
}

type T2 struct {
	S string
}

func (t *T2) M() {
	if t == nil {
		fmt.Println("<nil>")
		return
	}
	fmt.Println(t.S)
}

func nilUnderlying() {
	var i I

	var t *T2
	i = t
	describe(i)
	i.M()

	i = &T2{"hello"}
	describe(i)
	i.M()
}

func describe2(i interface{}) {
	fmt.Printf("(%v, %T)\n", i, i)
}

func emptyInterface() {
	var i interface{}
	describe2(i)

	i = 42
	describe2(i)

	i = "hello"
	describe2(i)
}

func typeAssertions() {
	var i interface{} = "hello"

	s := i.(string)
	fmt.Println(s)

	s, ok := i.(string)
	fmt.Println(s, ok)

	f, ok := i.(float64)
	fmt.Println(f, ok)

	//f = i.(float64) // panic
	//fmt.Println(f)
}

func do(i interface{}) {
	switch v := i.(type) {
	case int:
		fmt.Printf("Twice %v is %v\n", v, v*2)
	case string:
		fmt.Printf("%q is %v bytes long\n", v, len(v))
	default:
		fmt.Printf("I don't know about type %T!\n", v)
	}
}

func typeSwitches() {
	do(21)
	do("hello")
	do(true)
}

type Person struct {
	Name string
	Age  int
}

func (p Person) String() string {
	return fmt.Sprintf("%v (%v years)", p.Name, p.Age)
}

func stringers() {
	fmt.Println("******** stringers")

	a := Person{"Arthur Dent", 42}
	z := Person{"Zaphod Beeblebrox", 9001}
	fmt.Println(a, z)
}

type IPAddr [4]byte

// TODO: Add a "String() string" method to IPAddr.
func (ip IPAddr) String() string {
	var ipSlice []string
	for _, value := range ip[:] {
		ipSlice = append(ipSlice, fmt.Sprintf("%v", value))
	}
	return strings.Join(ipSlice, ".")
	//return fmt.Sprintf("%v.%v.%v.%v", ip[0], ip[1], ip[2], ip[3])
}

type MyError struct {
	When time.Time
	What string
}

func (e *MyError) Error() string {
	return fmt.Sprintf("at %v, %s",
		e.When, e.What)
}

func run() error {
	return &MyError{
		time.Now(),
		"it didn't work",
	}
}

func errors() {
	fmt.Println("******** errors")

	if err := run(); err != nil {
		fmt.Println(err)
	}
}

type ErrNegativeSqrt float64

func (e ErrNegativeSqrt) Error() string {
	var v float64 = float64(e)
	return fmt.Sprintf("cannot Sqrt negative number: %v", v)
}

func Sqrt(x float64) (float64, error) {
	if x < 0 {
		var e ErrNegativeSqrt = ErrNegativeSqrt(x)
		return 0, e
	}

	var lastZ float64 = x
	z := x / 2.0
	for math.Abs(x*1.0e-12) < math.Abs(z-lastZ) {
		lastZ = z
		z -= (z*z - x) / (2 * z)
	}
	return z, nil
}

func errorsExercise() {
	fmt.Println("******** errorsExercise")

	fmt.Println(Sqrt(2))
	fmt.Println(Sqrt(-2))
}

func readers() {
	fmt.Println("******** readers")

	r := strings.NewReader("Hello, Reader!")

	b := make([]byte, 8)
	for {
		n, err := r.Read(b)
		fmt.Printf("n = %v err = %v b = %v\n", n, err, b)
		fmt.Printf("b[:n] = %q\n", b[:n])
		if err == io.EOF {
			break
		}
	}
}

type MyReader struct{}

// TODO: Add a Read([]byte) (int, error) method to MyReader.
func (r MyReader) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = 'A'
	}
	return len(b), nil
}

func readersExercise() {
	fmt.Println("******** readersExercise")

	reader.Validate(MyReader{})
}

var (
	input  = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	output = []byte("NOPQRSTUVWXYZABCDEFGHIJKLMnopqrstuvwxyzabcdefghijklm")
)

type rot13Reader struct {
	r io.Reader
}

func Rot13(b byte) byte {
	for i, value := range input {
		if b == value {
			return output[i]
		}
	}
	return b
}

func (rot13 rot13Reader) Read(b []byte) (n int, err error) {
	n, err = rot13.r.Read(b)

	for i := range b[:n] {
		current := b[i]
		unrotated := Rot13(current)
		b[i] = unrotated
	}

	return
}

func rot13ReaderExercise() {
	fmt.Println("******** rot13ReaderExercise")

	s := strings.NewReader("Lbh penpxrq gur pbqr!")
	r := rot13Reader{s}
	io.Copy(os.Stdout, &r)
	fmt.Println("")
}

func images() {
	fmt.Println("******** images")

	m := image.NewRGBA(image.Rect(0, 0, 100, 100))
	fmt.Println(m.Bounds())
	fmt.Println(m.At(0, 0).RGBA())
}

type Image struct {
	X         int
	Y         int
	W         int
	H         int
	Rectangle image.Rectangle
}

func (img Image) ColorModel() color.Model {
	return color.RGBAModel
}

func (img Image) Bounds() image.Rectangle {
	img.Rectangle = image.Rect(0, 0, 200, 200)
	return img.Rectangle
}

func (img Image) At(x, y int) color.Color {
	//var v = uint8((x + y) / 2)
	//var v = uint8(x * y)
	//var v = uint8(x ^ y)
	var v = uint8(x - y)
	color := color.RGBA{v, v, 255, 255}
	return color
}

func imagesExercise() {
	fmt.Println("******** imagesExercise")

	m := Image{}
	pic.ShowImage(m)
}

func main() {
	methods()
	methods2()
	pointerReceivers()
	pointerIndirection()
	interfaces()
	var i I = &T{"hello"}
	i.M()
	interfaceValues()
	nilUnderlying()
	emptyInterface()
	typeAssertions()
	typeSwitches()
	stringers()

	hosts := map[string]IPAddr{
		"loopback":  {127, 0, 0, 1},
		"googleDNS": {8, 8, 8, 8},
	}
	for name, ip := range hosts {
		fmt.Printf("%v: %v\n", name, ip)
	}

	errors()
	errorsExercise()
	readers()
	readersExercise()
	rot13ReaderExercise()
	images()
	imagesExercise()
}
