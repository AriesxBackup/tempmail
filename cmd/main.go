package main

import (
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
	email := flag.String("email", "", "Email address to login")
	password := flag.String("password", "", "Password for the email account")
	fetch := flag.Bool("fetch", false, "Fetch all messages from inbox")
	domains := flag.Bool("domains", false, "List available email domains")
	monitor := flag.Int("monitor", 0, "Monitor inbox for N minutes")
	deleteAccount := flag.Bool("delete", false, "Delete the account after operations")

	flag.Parse()

	client := tempmail.NewClient()

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
			c := tempmail.NewClient()
			account, err := c.CreateAccount("")
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

	if *email != "" && *password != "" {
		fmt.Printf("Logging into %s...\n", *email)
		if err := client.Login(*email, *password); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Logged in successfully!")
	} else if *newAccount || (*email == "" && *password == "") {
		fmt.Println("Creating new temporary email account...")
		account, err := client.CreateAccount("")
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

		startTime := time.Now()
		duration := time.Duration(*monitor) * time.Minute
		checkCount := 0

		for time.Since(startTime) < duration {
			checkCount++
			messages, err := client.GetMessages(1)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				break
			}

			fmt.Printf("Check #%d - %d message(s)\n", checkCount, len(messages))

			if len(messages) > 0 {
				displayMessages(messages)
			}

			remaining := duration - time.Since(startTime)
			if remaining > 30*time.Second {
				fmt.Printf("Next check in 30s (remaining: %dm %ds)\n\n",
					int(remaining.Minutes()), int(remaining.Seconds())%60)
				time.Sleep(30 * time.Second)
			} else {
				break
			}
		}

		finalMessages, _ := client.GetMessages(1)
		fmt.Printf("\nFinal: %d total message(s)\n", len(finalMessages))
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
		fmt.Printf("    Date: %s\n", msg.CreatedAt)
		if msg.Seen {
			fmt.Println("    Status: Read")
		} else {
			fmt.Println("    Status: Unread")
		}
	}
	fmt.Println(strings.Repeat("=", 60))
}
