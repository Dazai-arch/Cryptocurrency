package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"crypto-portfolio-tracker/auth"
	"crypto-portfolio-tracker/portfolio"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n1. Signup\n2. Login\n3. Exit")
		fmt.Print("Choose option: ")
		choiceStr, _ := reader.ReadString('\n')
		choiceStr = strings.TrimSpace(choiceStr)
		choice, err := strconv.Atoi(choiceStr)

		if err != nil {
			fmt.Println("Invalid choice")
			continue
		}

		switch choice {
		case 1:
			fmt.Print("Enter Email: ")
			email, _ := reader.ReadString('\n')
			email = strings.TrimSpace(email)

			fmt.Print("Enter Password: ")
			password, _ := reader.ReadString('\n')
			password = strings.TrimSpace(password)

			auth.Signup(email, password, reader)

		case 2:
			fmt.Print("Enter Email: ")
			email, _ := reader.ReadString('\n')
			email = strings.TrimSpace(email)

			fmt.Print("Enter Password: ")
			password, _ := reader.ReadString('\n')
			password = strings.TrimSpace(password)

			if auth.Login(email, password) {
				fmt.Println("Welcome! Showing your portfolio:")
				portfolio.ShowPortfolio()
			}

		case 3:
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}
