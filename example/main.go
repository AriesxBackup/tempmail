package main

import (
	"fmt"
	"log"

	"tempmail/pkg/tempmail"
)

func main() {
	client := tempmail.NewClient()

	account, err := client.CreateAccount("")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Email: %s\n", account.Address)
	fmt.Printf("Password: %s\n\n", account.Password)

	messages, err := client.GetMessages(1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Messages: %d\n", len(messages))

	for _, msg := range messages {
		fmt.Printf("\nFrom: %s\n", msg.From.Address)
		fmt.Printf("Subject: %s\n", msg.Subject)
		fmt.Printf("Body: %s\n", msg.Text)
	}
}
