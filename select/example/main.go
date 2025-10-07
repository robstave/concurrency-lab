package main

import "fmt"

func main() {

	ch := make(chan string, 1)

	select {
	case msg := <-ch:
		fmt.Println("received:", msg)
	default:
		fmt.Println("no message received yet") // this one runs
	}

	ch <- "hello"

	select {
	case msg := <-ch:
		fmt.Println("received:", msg) // this one runs
	default:
		fmt.Println("no message received")
	}

	ch2 := make(chan string, 1)
	ch2 <- "first"

	select {
	case ch2 <- "second":
		fmt.Println("sent second")
	default:
		fmt.Println("channel full, couldn’t send") // this one runs
	}

	fmt.Println("received:", <-ch2)
	select {
	case ch2 <- "second":
		fmt.Println("sent second") // this one runs
	default:
		fmt.Println("channel full2 , couldn’t send")
	}

	fmt.Println("received:", <-ch2)

}
