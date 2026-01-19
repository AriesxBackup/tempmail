package main

import (
	"fmt"
	"log"

	"tempmail/pkg/tempmail"
)

func main() {
	// Create client
	client := tempmail.NewClient()

	// Create account (auto-generates password)
	account, err := client.CreateAccount("")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Email: %s\n", account.Address)
	fmt.Printf("Password: %s\n\n", account.Password)

	// Fetch all messages
	messages, err := client.GetMessages(1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Messages: %d\n", len(messages))

	// Read each message
	for _, msg := range messages {
		fmt.Printf("\nFrom: %s\n", msg.From.Address)
		fmt.Printf("Subject: %s\n", msg.Subject)
		fmt.Printf("Body: %s\n", msg.Text)
	}

	// Wait for new message (uncomment to use)
	// fmt.Println("\nWaiting for new message...")
	// newMsg, _ := client.WaitForMessage()
	// fmt.Printf("New message: %s\n", newMsg.Subject)

	// Delete account (optional)
	// client.DeleteAccount()
}
