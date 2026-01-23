package email

import (
	"fmt"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

func SendOTP(email, otp string) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	from := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")

	if from == "" || password == "" {
		fmt.Println("Email or password not set in .env file")
		return
	}

	msg := "From: " + from + "\n" +
		"To: " + email + "\n" +
		"Subject: Crypto Tracker OTP Verification\n\n" +
		"Your OTP is: " + otp

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")

	err = smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		from,
		[]string{email},
		[]byte(msg),
	)

	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}

	fmt.Println("OTP sent to", email)
}
