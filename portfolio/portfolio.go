package portfolio

import "fmt"

func ShowPortfolio() {
	var coins [3]string = [3]string{"Bitcoin", "Ethereum", "Ripple"}
	fmt.Println("Fixed array of coins:", coins)

	portfolio := []string{}
	portfolio = append(portfolio, "Bitcoin")
	portfolio = append(portfolio, "Ethereum")
	portfolio = append(portfolio, "Cardano")
	fmt.Println("Slice portfolio (dynamic):", portfolio)

	quantities := [3]float64{1.5, 2.0, 100}
	fmt.Println("Quantities array:", quantities)

	sliceQuantities := []float64{}
	sliceQuantities = append(sliceQuantities, 0.75)
	sliceQuantities = append(sliceQuantities, 1.2)
	sliceQuantities = append(sliceQuantities, 50.0)
	fmt.Println("Quantities slice:", sliceQuantities)

	fmt.Println("Your portfolio:")
	for i, coin := range portfolio {
		fmt.Printf("%s: %.2f coins\n", coin, sliceQuantities[i])
	}
}
