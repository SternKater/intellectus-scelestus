package main

import "fmt"

type Speaker interface {
	Speak() string
}

type Dog struct {
	Name string
}

func (d Dog) Speak() string {
	return "Woof, my name is " + d.Name
}

func main() {
	var s Speaker
	s = Dog{Name: "Бобик"}

	fmt.Println(s.Speak())
}
