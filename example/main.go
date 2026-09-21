package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"tempmail/pkg/tempmail"
)

func main() {
	client := tempmail.NewClient(tempmail.WithAPIKey(os.Getenv("SMTPDEV_API_KEY")))

	account, err := client.CreateAccount("")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Email: %s\n", account.Address)
	fmt.Printf("Password: %s\n\n", account.Password)
	fmt.Println("Waiting for a message...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	msg, err := client.WaitForMessage(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nFrom: %s\n", msg.From.Address)
	fmt.Printf("Subject: %s\n", msg.Subject)
	fmt.Printf("Body: %s\n", msg.Text)

	if err := client.DeleteAccount(); err != nil {
		log.Fatal(err)
	}
}
