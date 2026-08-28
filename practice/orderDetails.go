package main

import "fmt"

// User struct
type User struct {
	Name  string
	Email string
}

// Order struct that embeds a User
type Order struct {
	ID     int
	User   User
	Amount float64
}

// Show order details
func (o Order) OrderDetails() {
	fmt.Println("===== Order Details =====")
	fmt.Printf("Order ID : %d\n", o.ID)
	fmt.Printf("User     : %s (%s)\n", o.User.Name, o.User.Email)
	fmt.Printf("Amount   : %.2f\n", o.Amount)
}

// Change order amount
func (o *Order) ChangeAmount(newAmount float64) {
	o.Amount = newAmount
}

func main() {
	// Create user
	user := User{
		Name:  "Kiran",
		Email: "kiran@example.com",
	}

	// Create order
	order := Order{
		ID:     101,
		User:   user,
		Amount: 1500.75,
	}

	// Before update
	order.OrderDetails()

	var shippingAmount float64 = 100.0
	fmt.Printf("Shipping Amount: %.2f\n", shippingAmount)

	// Update amount (Packing + Shipping)
	order.ChangeAmount(order.Amount + shippingAmount)

	fmt.Printf("\nAfter Updating Amount (Because of Packing/Shipping Amount): %.2f\n", order.Amount)
	order.OrderDetails()
}
