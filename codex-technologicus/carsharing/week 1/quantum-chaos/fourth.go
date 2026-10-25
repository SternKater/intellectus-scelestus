package main

import "fmt"

func calculateStats() (int, int) {
	return 100, 50
}

func main() {
	score, _ := calculateStats()

	fmt.Println("Score:", score)
}
