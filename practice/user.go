package main

import "fmt"

type User struct {
	Name string
	Age  int
}

func (u User) Greet() {
	fmt.Println("Hello", u.Name)
}

func (u *User) Birthday() {
	u.Age++
}

func main() {
	u := User{Name: "Kiran", Age: 25}

	u.Greet()    // "Hello Kiran"
	u.Birthday() // Age = 26
	fmt.Println(u.Age)
}
