# tempmail

A Go wrapper for the [mail.tm](https://mail.tm) API. Based on [pymailtm](https://github.com/CarloDePieri/pymailtm).

## Installation

```bash
go get github.com/yourusername/tempmail
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "tempmail/pkg/tempmail"
)

func main() {
    client := tempmail.NewClient()
    
    // Create account
    account, err := client.CreateAccount("")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Email: %s\n", account.Address)
    fmt.Printf("Password: %s\n", account.Password)
    
    // Fetch messages
    messages, _ := client.GetMessages(1)
    fmt.Printf("Messages: %d\n", len(messages))
}
```

## CLI Usage

```bash
# Build
go build -o tempmail ./cmd

# Create new account
./tempmail --new

# Create and fetch messages
./tempmail --new --fetch

# Create multiple accounts
./tempmail --count 5

# Login to existing account
./tempmail --email user@domain.com --password pass --fetch

# List available domains
./tempmail --domains

# Monitor inbox
./tempmail --new --monitor 10

# Delete account after use
./tempmail --new --fetch --delete
```

## API Reference

### Create Client

```go
client := tempmail.NewClient()

// With options
client := tempmail.NewClient(
    tempmail.WithAPIAddress("https://api.mail.tm"),
    tempmail.WithTimeout(60 * time.Second),
)
```

### Create Account

```go
// Auto-generate password
account, err := client.CreateAccount("")

// Custom password
account, err := client.CreateAccount("mypassword")
```

### Login

```go
err := client.Login("email@domain.com", "password")
```

### Get Messages

```go
messages, err := client.GetMessages(1) // page number
```

### Get Single Message

```go
msg, err := client.GetMessage("message-id")
```

### Mark Message as Seen

```go
err := client.MarkMessageSeen("message-id")
```

### Delete Message

```go
err := client.DeleteMessage("message-id")
```

### Delete Account

```go
err := client.DeleteAccount()
```

### Wait for New Message

```go
msg, err := client.WaitForMessage()
```

## Project Structure

```
.
├── cmd/
│   └── main.go           # CLI application
├── example/
│   └── main.go           # Example usage
├── pkg/
│   └── tempmail/
│       ├── client.go     # Main client
│       ├── client_test.go
│       ├── errors.go     # Error types
│       ├── models.go     # Data structures
│       ├── utils.go      # Helpers
│       └── utils_test.go
├── go.mod
├── .gitignore
└── README.md
```

## Running Tests

```bash
go test ./pkg/tempmail/... -v
```

## License

MIT
