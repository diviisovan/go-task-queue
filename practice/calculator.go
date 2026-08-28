package main

import (
	"fmt"
)

func main() {
	var a, b float64
	var op string

	fmt.Print("Enter first number: ")
	fmt.Scanln(&a)

	fmt.Print("Enter operator (+, -, *, /): ")
	fmt.Scanln(&op)

	fmt.Print("Enter second number: ")
	fmt.Scanln(&b)

	switch op {
	case "+":
		fmt.Println("Result:", a+b)
	case "-":
		fmt.Println("Result:", a-b)
	case "*":
		fmt.Println("Result:", a*b)
	case "/":
		if b != 0 {
			fmt.Println("Result:", a/b)
		} else {
			fmt.Println("Error: Cannot divide by zero")
		}
	default:
		fmt.Println("Invalid operator")
	}
}
