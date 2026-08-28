package main

import (
	"fmt"
)

type Product struct {
	Name  string
	Price float64
}

// Method 1: Show product details (value receiver)
func (p Product) ShowDetails() {
	fmt.Printf("Product: %s | Price: %.2f\n", p.Name, p.Price)
}

// Method 2: Apply discount (pointer receiver, modifies original)
func (p *Product) ApplyDiscount(percent float64) {
	discount := p.Price * (percent / 100)
	p.Price = p.Price - discount
}

func main() {
	// Create a product
	p := Product{Name: "Laptop", Price: 50000}

	// Show initial details
	p.ShowDetails()

	// Apply 10% discount
	p.ApplyDiscount(10)

	// Show updated details
	p.ShowDetails()
}
