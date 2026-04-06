package main

import (
	"fmt"
	"sync"
)

type Singleton struct {
	data string
}

var instance *Singleton
var once sync.Once

func GetInstance() *Singleton {
	once.Do(func() {
		fmt.Println("Creating Singleton instance")
		instance = &Singleton{data: "I'm the only one!"}
	})
	return instance
}

func main() {
	for i := 0; i < 5; i++ {
		go func() {
			fmt.Printf("%p\n", GetInstance())
		}()
	}

	// Wait for goroutines to finish
	fmt.Scanln()
}
