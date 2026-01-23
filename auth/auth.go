package auth

import (
	"bufio"
	"context"
	"crypto-portfolio-tracker/db"
	"crypto-portfolio-tracker/email"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email    string `bson:"email"`
	Password string `bson:"password"`
	Verified bool   `bson:"verified"`
	OTP      string `bson:"otp"`
}

func generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func hashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashed)
}

func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func Signup(userEmail, password string, reader *bufio.Reader) {
	database := db.ConnectDatabase()
	collection := database.Collection("users")

	var existingUser User
	err := collection.FindOne(
		context.TODO(),
		bson.M{"email": userEmail},
	).Decode(&existingUser)

	if err == nil {
		fmt.Println("Account already exists with this email.")
		return
	}

	if err != mongo.ErrNoDocuments {
		fmt.Println("Database error:", err)
		return
	}

	otp := generateOTP()
	email.SendOTP(userEmail, otp)

	fmt.Print("Enter OTP: ")
	inputOTP, _ := reader.ReadString('\n')
	inputOTP = strings.TrimSpace(inputOTP)

	if inputOTP != otp {
		fmt.Println("Invalid OTP. Signup failed.")
		return
	}

	user := User{
		Email:    userEmail,
		Password: hashPassword(password),
		Verified: true,
		OTP:      "",
	}

	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		fmt.Println("Signup failed:", err)
		return
	}

	fmt.Println("Signup successful!")
}

func Login(email, password string) bool {
	database := db.ConnectDatabase()
	var u User

	database.Collection("users").FindOne(
		context.TODO(),
		map[string]string{"email": email},
	).Decode(&u)

	if !u.Verified {
		fmt.Println("Email not verified")
		return false
	}

	if checkPassword(u.Password, password) {
		fmt.Println("Login successful!")
		return true
	}
	return false
}
