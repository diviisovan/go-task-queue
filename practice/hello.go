package main

import "fmt"

func main() {
	fmt.Println("Hello Go!")
	user := User{Name: "Alice", Age: 30}
	user.Greet()
}

type User struct {
	Name string
	Age  int
}

func (u User) Greet() {
	fmt.Println("Hi", u.Name)
}
