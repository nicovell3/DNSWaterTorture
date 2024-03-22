package main

import (
	"fmt"
	"log"
	"testing"
)

func TestRandomString(t *testing.T) {
	fmt.Println("Creating many random strings")
	fmt.Print("\033[s")
	for i := 0; i < 10000; i++ {
		fmt.Print("\033[u\033[K")
		fmt.Printf("%d %d\n", i, len(randomString(subdomainLength)))
	}
	log.Println("Created random strings successfully")
}
