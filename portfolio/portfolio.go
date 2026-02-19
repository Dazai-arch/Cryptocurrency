package portfolio

import (
	"context"
	"crypto-portfolio-tracker/api"
	"crypto-portfolio-tracker/db"
	customerrors "crypto-portfolio-tracker/errors"
	"crypto-portfolio-tracker/models"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddMultipleHoldings(userEmail string, holdings ...models.Holding) error {
	if len(holdings) == 0 {
		return customerrors.ErrEmptyHoldings
	}

	database, err := db.ConnectDatabase()
	if err != nil {
		return customerrors.NewDatabaseError("connect", "portfolios", err)
	}

	collection := database.Collection("portfolios")

	for _, holding := range holdings {
		if holding.Quantity <= 0 {
			return customerrors.NewValidationError("quantity", holding.Quantity, customerrors.ErrInvalidQuantity)
		}
		if holding.BuyPrice <= 0 {
			return customerrors.NewValidationError("buy_price", holding.BuyPrice, customerrors.ErrInvalidPrice)
		}

		holding.AddedAt = time.Now()

		filter := bson.M{"user_email": userEmail, "holdings.coin_id": holding.CoinID}
		update := bson.M{
			"$inc": bson.M{"holdings.$.quantity": holding.Quantity},
			"$set": bson.M{"updated_at": time.Now()},
		}

		result, err := collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return customerrors.NewDatabaseError("update", "portfolios", err)
		}

		if result.MatchedCount == 0 {
			filter = bson.M{"user_email": userEmail}
			update = bson.M{
				"$push": bson.M{"holdings": holding},
				"$set":  bson.M{"updated_at": time.Now()},
			}

			_, err = collection.UpdateOne(
				context.TODO(),
				filter,
				update,
				options.Update().SetUpsert(true),
			)
			if err != nil {
				return customerrors.NewPortfolioError("add holding", holding.CoinID, err)
			}
		}
	}

	return nil
}

func GetPortfolio(userEmail string) (*models.Portfolio, error) {
	database, err := db.ConnectDatabase()
	if err != nil {
		return nil, customerrors.NewDatabaseError("connect", "portfolios", err)
	}

	var portfolio models.Portfolio
	collection := database.Collection("portfolios")

	err = collection.FindOne(
		context.TODO(),
		bson.M{"user_email": userEmail},
	).Decode(&portfolio)

	if err == mongo.ErrNoDocuments {
		return &models.Portfolio{
			UserEmail: userEmail,
			Holdings:  []models.Holding{},
			UpdatedAt: time.Now(),
		}, nil
	}

	if err != nil {
		return nil, customerrors.NewDatabaseError("fetch", "portfolios", err)
	}

	return &portfolio, nil
}

func CalculateTotalValue(portfolio *models.Portfolio, apiClient api.CryptoApi) (float64, error) {
	if len(portfolio.Holdings) == 0 {
		return 0, nil
	}

	coinIDs := make([]string, len(portfolio.Holdings))
	for i, holding := range portfolio.Holdings {
		coinIDs[i] = holding.CoinID
	}

	prices, err := apiClient.FetchMultiplePrices(coinIDs...)
	if err != nil {
		return 0, customerrors.NewPortfolioError("calculate total value", "", err)
	}

	var totalValue float64
	for _, holding := range portfolio.Holdings {
		currentPrice, ok := prices[holding.CoinID]
		if !ok {
			return 0, customerrors.NewPortfolioError("calculate total value", holding.CoinID, customerrors.ErrPriceNotAvailable)
		}
		totalValue += currentPrice * holding.Quantity
	}

	return totalValue, nil
}

func CalculateProfitLoss(portfolio *models.Portfolio, apiClient api.CryptoApi, coinIDs ...string) (map[string]float64, error) {
	profitLoss := make(map[string]float64)

	if len(coinIDs) == 0 {
		for _, holding := range portfolio.Holdings {
			coinIDs = append(coinIDs, holding.CoinID)
		}
	}

	if len(coinIDs) == 0 {
		return profitLoss, customerrors.ErrEmptyPortfolio
	}

	prices, err := apiClient.FetchMultiplePrices(coinIDs...)
	if err != nil {
		return nil, customerrors.NewPortfolioError("calculate profit/loss", "", err)
	}

	holdingsMap := make(map[string]models.Holding)
	for _, holding := range portfolio.Holdings {
		holdingsMap[holding.CoinID] = holding
	}

	for _, coinID := range coinIDs {
		holding, exists := holdingsMap[coinID]
		if !exists {
			continue
		}

		currentPrice, ok := prices[coinID]
		if !ok {
			return nil, customerrors.NewPortfolioError("calculate profit/loss", coinID, customerrors.ErrPriceNotAvailable)
		}

		invested := holding.BuyPrice * holding.Quantity
		current := currentPrice * holding.Quantity
		profitLoss[coinID] = current - invested
	}

	return profitLoss, nil
}

func DisplayPortfolio(userEmail string, apiClient api.CryptoApi) error {
	portfolio, err := GetPortfolio(userEmail)
	if err != nil {
		return customerrors.NewPortfolioError("display portfolio", "", err)
	}

	if len(portfolio.Holdings) == 0 {
		fmt.Println("Your portfolio is empty.")
		return nil
	}

	fmt.Println("\n========== YOUR PORTFOLIO ==========")

	for _, holding := range portfolio.Holdings {
		currentPrice, err := apiClient.FetchPrice(holding.CoinID)
		if err != nil {
			fmt.Printf("Warning: Could not fetch price for %s: %v\n", holding.CoinName, err)
			continue
		}

		currentValue := currentPrice * holding.Quantity
		invested := holding.BuyPrice * holding.Quantity
		profitLoss := currentValue - invested
		profitLossPercent := (profitLoss / invested) * 100

		fmt.Printf("\nCoin: %s (%s)\n", holding.CoinName, holding.CoinID)
		fmt.Printf("  Quantity: %.4f\n", holding.Quantity)
		fmt.Printf("  Buy Price: $%.2f\n", holding.BuyPrice)
		fmt.Printf("  Current Price: $%.2f\n", currentPrice)
		fmt.Printf("  Current Value: $%.2f\n", currentValue)
		fmt.Printf("  Profit/Loss: $%.2f (%.2f%%)\n", profitLoss, profitLossPercent)
	}

	totalValue, err := CalculateTotalValue(portfolio, apiClient)
	if err != nil {
		return customerrors.NewPortfolioError("calculate total value", "", err)
	}

	fmt.Printf("\n====================================\n")
	fmt.Printf("Total Portfolio Value: $%.2f\n", totalValue)
	fmt.Printf("====================================\n\n")

	return nil
}
