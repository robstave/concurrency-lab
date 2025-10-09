package main

import (
	"fmt"
	"time"

	"example.com/mapmutex/regularintmap"
)

func main() {

	m := regularintmap.NewRegularIntMap()

	// Store some values concurrently
	for i := range 5 {
		go func(i int) {
			m.Store(i, i*i)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	// Load values
	for i := range 5 {
		if v, ok := m.Load(i); ok {
			fmt.Printf("key=%d value=%d\n", i, v)
		}
	}

	// Delete a key then show it's gone
	m.Delete(2)
	if _, ok := m.Load(2); !ok {
		fmt.Println("key=2 deleted")
	}
}
