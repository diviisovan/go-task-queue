package main

import "fmt"

type Employee struct {
	Name   string
	Salary int
}

func (e *Employee) Hike(percent int) {
	increase := float64(e.Salary) * (float64(percent) / 100)
	e.Salary = e.Salary + int(increase)
}

func (e *Employee) Display() {
	fmt.Printf("Employee: %s | Salary: %d\n", e.Name, e.Salary)
}

func main() {
	var percentage int
	var salary int

	fmt.Print("Enter salary: ")
	fmt.Scanln(&salary)

	employee := Employee{Name: "John", Salary: salary}

	fmt.Printf("Salary before hike: %d\n", employee.Salary)

	fmt.Print("Enter percentage to hike: ")
	fmt.Scanln(&percentage)

	employee.Hike(percentage)

	fmt.Printf("Salary after hike: %d\n", employee.Salary)

	employee.Display()
}
