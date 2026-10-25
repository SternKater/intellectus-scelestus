package second

import "fmt"

// Базовая структура
type Vehicle struct {
	Speed int
}

func (v Vehicle) Move() {
	fmt.Printf("Speed %d kmph\n", v.Speed)
}

type ElectricCar struct {
	Vehicle
	BatteryCap int
}

func main() {
	tesla := ElectricCar{
		Vehicle:    Vehicle{Speed: 120},
		BatteryCap: 100,
	}

	tesla.Move()

	fmt.Println("Current:", tesla.Speed)
}
