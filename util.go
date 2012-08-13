package main

import "math/rand"

func randRange(mean int, sigma int) int {
	return mean + (rand.Int() % sigma) - (rand.Int() % sigma)
}