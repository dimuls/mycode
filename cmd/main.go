package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var p float64
	fmt.Scanf("%g", &p)

	if rand.Float64() <= p {
		fmt.Print("ok")
	}
}
