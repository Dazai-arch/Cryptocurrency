package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDatabase() *mongo.Database {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	uri := os.Getenv("MONGO_URI")

	if uri == "" {
		log.Fatal("MONGO_URI not set in environment")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
		return nil
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		fmt.Println("MongoDB Ping Failed:", err)
		return nil
	}

	//fmt.Println("MongoDB Connection Established")
	return client.Database("crypto_tracker")
}

// func main() {
// 	db := ConnectDatabase()
// 	if db != nil {
// 		fmt.Println("You can now use the database:", db.Name())
// 	} else {
// 		fmt.Println("Failed to connect to database.")
// 	}
// }
