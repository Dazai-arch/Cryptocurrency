package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"crypto-portfolio-tracker/api"
	"crypto-portfolio-tracker/auth"
	customerrors "crypto-portfolio-tracker/errors"
	"crypto-portfolio-tracker/models"
	"crypto-portfolio-tracker/portfolio"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	cryptoAPI, err := api.NewCoinGecko()
	if err != nil {
		fmt.Printf("Failed to initialize CoinGecko API: %v\n", err)
		return
	}

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
				handlePortfolioMenu(email, cryptoAPI, reader)
			}

		case 3:
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func handlePortfolioMenu(userEmail string, cryptoAPI api.CryptoApi, reader *bufio.Reader) {
	for {
		fmt.Println("\n=== Portfolio Menu ===")
		fmt.Println("1. View Portfolio")
		fmt.Println("2. Add Holdings")
		fmt.Println("3. Add Multiple Holdings")
		fmt.Println("4. Calculate Total Value")
		fmt.Println("5. Calculate Profit/Loss Value")
		fmt.Println("6. LogOut")
		fmt.Print("Enter The Option: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		option, err := strconv.Atoi(choice)

		if err != nil {
			fmt.Println("Invalid Choice Try Again!!")
			continue
		}

		switch option {
		case 1:
			if err := portfolio.DisplayPortfolio(userEmail, cryptoAPI); err != nil {
				fmt.Printf("Error displaying portfolio: %v\n", err)
			}

		case 2:
			addSingleHolding(userEmail, reader)

		case 3:
			addMultipleHoldings(userEmail, reader)

		case 4:
			calculateTotal(userEmail, cryptoAPI)

		case 5:
			calculateProfitLoss(userEmail, cryptoAPI, reader)

		case 6:
			fmt.Println("Logging Out")
			return
		default:
			fmt.Println("Invalid Choice")
		}
	}
}

func addMultipleHoldings(userEmail string, reader *bufio.Reader) {
	fmt.Print("How many holdings do you want to add? ")
	countStr, _ := reader.ReadString('\n')
	count, err := strconv.Atoi(strings.TrimSpace(countStr))
	if err != nil || count <= 0 {
		fmt.Println("Invalid count")
		return
	}

	holdings := make([]models.Holding, count)
	for i := 0; i < count; i++ {
		fmt.Printf("\n--- Holding %d ---\n", i+1)

		fmt.Print("Coin ID: ")
		coinID, _ := reader.ReadString('\n')

		fmt.Print("Coin Name: ")
		coinName, _ := reader.ReadString('\n')

		fmt.Print("Quantity: ")
		quantityStr, _ := reader.ReadString('\n')
		quantity, _ := strconv.ParseFloat(strings.TrimSpace(quantityStr), 64)

		fmt.Print("Buy Price: ")
		priceStr, _ := reader.ReadString('\n')
		buyPrice, _ := strconv.ParseFloat(strings.TrimSpace(priceStr), 64)

		holdings[i] = models.Holding{
			CoinID:   strings.TrimSpace(coinID),
			CoinName: strings.TrimSpace(coinName),
			Quantity: quantity,
			BuyPrice: buyPrice,
		}
	}
	if err := portfolio.AddMultipleHoldings(userEmail, holdings...); err != nil {
		fmt.Printf("Error adding holdings: %v\n", err)
		return
	}

	fmt.Println("All holdings added successfully!")
}

func calculateTotal(userEmail string, cryptoAPI api.CryptoApi) {
	p, err := portfolio.GetPortfolio(userEmail)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	total, err := portfolio.CalculateTotalValue(p, cryptoAPI)
	if err != nil {
		fmt.Printf("Error calculating total: %v\n", err)
		return
	}

	fmt.Printf("\nTotal Portfolio Value: $%.2f\n", total)
}

func calculateProfitLoss(userEmail string, cryptoAPI api.CryptoApi, reader *bufio.Reader) {
	fmt.Println("\n⏳ Loading your portfolio...")
	p, err := portfolio.GetPortfolio(userEmail)
	if err != nil {
		var dbErr *customerrors.DatabaseError
		if errors.As(err, &dbErr) {
			fmt.Printf("Database Error: %v\n", dbErr)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if len(p.Holdings) == 0 {
		fmt.Println("📭 Your portfolio is empty. Add some holdings first!")
		return
	}

	fmt.Print("Calculate for specific coins? (y/n): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(strings.ToLower(choice))

	var profitLoss map[string]float64

	if choice == "y" {
		fmt.Print("Enter coin IDs (comma-separated): ")
		coinsStr, _ := reader.ReadString('\n')
		coinIDs := strings.Split(strings.TrimSpace(coinsStr), ",")

		for i := range coinIDs {
			coinIDs[i] = strings.TrimSpace(coinIDs[i])
		}

		fmt.Println("\nCalculating profit/loss...")
		profitLoss, err = portfolio.CalculateProfitLoss(p, cryptoAPI, coinIDs...)
	} else {
		fmt.Println("\nCalculating profit/loss for all holdings...")
		profitLoss, err = portfolio.CalculateProfitLoss(p, cryptoAPI)
	}

	if err != nil {
		if errors.Is(err, customerrors.ErrRateLimitExceeded) {
			fmt.Println("Rate limit exceeded!")
			fmt.Println("Please wait 5-10 seconds and try again.")
		} else if errors.Is(err, customerrors.ErrPriceNotAvailable) {
			fmt.Println("Price data not available for one or more coins")
		} else {
			var apiErr *customerrors.APIError
			var portfolioErr *customerrors.PortfolioError

			if errors.As(err, &apiErr) {
				fmt.Printf("API Error [%d]: %v\n", apiErr.StatusCode, apiErr)
			} else if errors.As(err, &portfolioErr) {
				fmt.Printf("Portfolio Error: %v\n", portfolioErr)
			} else {
				fmt.Printf("Error: %v\n", err)
			}
		}
		return
	}

	if len(profitLoss) == 0 {
		fmt.Println("No profit/loss data available.")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("PROFIT/LOSS ANALYSIS")
	fmt.Println(strings.Repeat("=", 60))

	var totalProfitLoss float64
	for coin, pl := range profitLoss {
		fmt.Printf(" %-20s: $%+.2f\n", coin, pl)
		totalProfitLoss += pl
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("TOTAL PROFIT/LOSS: $%+.2f\n", totalProfitLoss)
	fmt.Println(strings.Repeat("=", 60))
}

func addSingleHolding(userEmail string, reader *bufio.Reader) {
	fmt.Print("Enter Coin ID (e.g., bitcoin): ")
	coinID, _ := reader.ReadString('\n')
	coinID = strings.TrimSpace(coinID)

	fmt.Print("Enter Coin Name (e.g., Bitcoin): ")
	coinName, _ := reader.ReadString('\n')
	coinName = strings.TrimSpace(coinName)

	fmt.Print("Enter Quantity: ")
	quantityStr, _ := reader.ReadString('\n')
	quantity, err := strconv.ParseFloat(strings.TrimSpace(quantityStr), 64)
	if err != nil {
		fmt.Println("Invalid quantity")
		return
	}

	fmt.Print("Enter Buy Price: ")
	priceStr, _ := reader.ReadString('\n')
	buyPrice, err := strconv.ParseFloat(strings.TrimSpace(priceStr), 64)
	if err != nil {
		fmt.Println("Invalid price")
		return
	}

	holding := models.Holding{
		CoinID:   coinID,
		CoinName: coinName,
		Quantity: quantity,
		BuyPrice: buyPrice,
	}

	if err := portfolio.AddMultipleHoldings(userEmail, holding); err != nil {
		if errors.Is(err, customerrors.ErrInvalidQuantity) {
			fmt.Println("Quantity must be greater than 0")
		} else if errors.Is(err, customerrors.ErrInvalidPrice) {
			fmt.Println("Buy price must be greater than 0")
		} else {
			fmt.Printf("Error adding holding: %v\n", err)
		}
		return
	}

	fmt.Println("Holding added successfully!")
}
