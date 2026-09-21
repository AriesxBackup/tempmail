package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"tempmail/pkg/tempmail"
)

func main() {
	newAccount := flag.Bool("new", false, "Create a new email account")
	count := flag.Int("count", 1, "Number of email accounts to create")
	email := flag.String("email", "", "Existing email address to use")
	password := flag.String("password", "", "Password for a new account (auto-generated if empty)")
	apiKey := flag.String("key", os.Getenv("SMTPDEV_API_KEY"), "smtp.dev API key (or set SMTPDEV_API_KEY)")
	fetch := flag.Bool("fetch", false, "Fetch all messages from inbox")
	domains := flag.Bool("domains", false, "List available email domains")
	monitor := flag.Int("monitor", 0, "Monitor inbox for N minutes")
	deleteAccount := flag.Bool("delete", false, "Delete the account after operations")

	flag.Parse()

	if *apiKey == "" {
		fmt.Println("Error: an API key is required (use --key or SMTPDEV_API_KEY). Get one at https://smtp.dev/tokens")
		os.Exit(1)
	}

	client := tempmail.NewClient(tempmail.WithAPIKey(*apiKey))

	if *domains {
		domainList, err := client.GetDomains()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Available email domains:")
		for _, d := range domainList {
			fmt.Printf("  - %s\n", d)
		}
		return
	}

	if *count > 1 {
		fmt.Printf("Creating %d temporary email accounts...\n\n", *count)
		var accounts []tempmail.Account

		for i := 0; i < *count; i++ {
			c := tempmail.NewClient(tempmail.WithAPIKey(*apiKey))
			account, err := c.CreateAccount(*password)
			if err != nil {
				fmt.Printf("Error creating account %d: %v\n\n", i+1, err)
				continue
			}
			accounts = append(accounts, *account)
			fmt.Printf("Account %d/%d created:\n", i+1, *count)
			fmt.Printf("   Email: %s\n", account.Address)
			fmt.Printf("   Password: %s\n\n", account.Password)
		}

		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Printf("Summary: %d/%d accounts created\n", len(accounts), *count)
		fmt.Println(strings.Repeat("=", 60))
		return
	}

	if *email != "" && !*newAccount {
		fmt.Printf("Looking up %s...\n", *email)
		if _, err := client.UseAccount(*email); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Account found!")
	} else {
		fmt.Println("Creating new temporary email account...")
		var (
			account *tempmail.Account
			err     error
		)
		if *email != "" {
			account, err = client.CreateAccountWithAddress(*email, *password)
		} else {
			account, err = client.CreateAccount(*password)
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Account created successfully!")
		fmt.Printf("Email: %s\n", account.Address)
		fmt.Printf("Password: %s\n", account.Password)
		fmt.Printf("Account ID: %s\n", account.ID)
	}

	if *fetch {
		fmt.Println("\nFetching all messages...")
		messages, err := client.GetMessages(1)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		displayMessages(messages)

		if len(messages) > 0 {
			fmt.Println("Latest Message Content:")
			fmt.Println(strings.Repeat("=", 60))
			msg := messages[0]
			fmt.Printf("From: %s\n", msg.From.Address)
			fmt.Printf("Subject: %s\n", msg.Subject)
			if msg.Text != "" {
				fmt.Printf("\n%s\n", msg.Text)
			} else {
				fmt.Println("\nNo text content")
			}
			fmt.Println(strings.Repeat("=", 60))
		}
	}

	if *monitor > 0 {
		fmt.Printf("\nMonitoring inbox for %d minutes...\n", *monitor)
		if client.GetAccount() != nil {
			fmt.Printf("Email: %s\n", client.GetAccount().Address)
		}
		fmt.Println("Press Ctrl+C to stop")

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*monitor)*time.Minute)
		defer cancel()

		received := 0
		for {
			msg, err := client.WaitForMessage(ctx)
			if err != nil {
				if !errors.Is(err, tempmail.ErrTimeout) {
					fmt.Printf("Error: %v\n", err)
				}
				break
			}
			received++
			displayMessages([]tempmail.Message{*msg})
		}

		fmt.Printf("\nFinal: %d new message(s) received\n", received)
	}

	if *deleteAccount && client.GetAccount() != nil {
		fmt.Printf("\nDeleting account %s...\n", client.GetAccount().Address)
		if err := client.DeleteAccount(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Account deleted!")
	}
}

func displayMessages(messages []tempmail.Message) {
	if len(messages) == 0 {
		fmt.Println("No messages found.")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("INBOX - %d message(s)\n", len(messages))
	fmt.Println(strings.Repeat("=", 60))

	for i, msg := range messages {
		fmt.Printf("\n[%d] %s\n", i+1, msg.Subject)
		fmt.Printf("    From: %s\n", msg.From.Address)
		fmt.Printf("    Date: %s\n", msg.Date)
		if msg.IsRead {
			fmt.Println("    Status: Read")
		} else {
			fmt.Println("    Status: Unread")
		}
	}
	fmt.Println(strings.Repeat("=", 60))
}
