package main

import "fmt"

type BankAccount struct {
	Owner   string
	Balance float64
}

func (b BankAccount) ShowBalance() float64 {
	return b.Balance
}

func (b *BankAccount) Deposit(amount float64) {
	b.Balance += amount
}

func (b *BankAccount) Withdraw(amount float64) {
	if b.Balance >= amount {
		b.Balance -= amount
	} else {
		fmt.Println("Insufficient balance")
	}
}

func main() {
	account := BankAccount{Owner: "Amit", Balance: 1000}
	fmt.Println("Current Balance =", account.ShowBalance())
	var amountDeposit float64
	fmt.Print("Enter amount to deposit: ")
	fmt.Scanln(&amountDeposit)
	account.Deposit(amountDeposit)
	fmt.Println("Post-Deposit Balance =", account.ShowBalance())
	var amountWithdraw float64
	fmt.Print("Enter amount to withdraw: ")
	fmt.Scanln(&amountWithdraw)
	account.Withdraw(amountWithdraw)
	fmt.Println("Post-Withdraw Balance =", account.ShowBalance())
}
