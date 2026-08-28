package main

import "fmt"

// Item struct
type Item struct {
	Name  string
	Price float64
}

// Cart struct (contains a slice of Items)
type Cart struct {
	Items []Item
}

// AddItem adds an item to the cart
func (c *Cart) AddItem(item Item) {
	c.Items = append(c.Items, item)
}

// TotalPrice calculates the total price of the cart
func (c *Cart) TotalPrice() float64 {
	total := 0.0
	for _, item := range c.Items {
		total += item.Price
	}
	return total
}

// ShowItems displays all items in the cart
func (c *Cart) ShowItems() {
	fmt.Println("\nItems in Cart:")
	for _, item := range c.Items {
		fmt.Printf("- %s : %.2f\n", item.Name, item.Price)
	}
	fmt.Printf("Total Price: %.2f\n", c.TotalPrice())
}

func main() {
	cart := Cart{}

	// Manually adding items
	cart.AddItem(Item{Name: "Pen", Price: 10})
	cart.AddItem(Item{Name: "Notebook", Price: 50})
	cart.AddItem(Item{Name: "Bag", Price: 300.50})

	// Show cart items and total
	cart.ShowItems()
}

// TASK 4 — Shopping Cart
// Requirements:

// Create:

// Item struct → Name, Price

// Cart struct → slice of items

// Methods:

// AddItem(item)

// TotalPrice()

// ShowItems()

// Example:

// Items:
// - Pen: 10
// - Book: 100
// Total: 110
