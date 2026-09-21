# tempmail

A Go wrapper for the [smtp.dev](https://smtp.dev/docs/api/) API ([OpenAPI spec](https://api.smtp.dev/docs.jsonopenapi)).

## Installation

```bash
go get github.com/AriesxBackup/tempmail
```

## Authentication

Every request needs an API key from [smtp.dev/tokens](https://smtp.dev/tokens), sent as the `X-API-KEY` header.

```go
client := tempmail.NewClient(tempmail.WithAPIKey(os.Getenv("SMTPDEV_API_KEY")))
```

## Quick Start

```go
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

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    msg, err := client.WaitForMessage(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(msg.Subject, msg.Text)
}
```

## CLI Usage

```bash
export SMTPDEV_API_KEY=smtplabs_...

go build -o tempmail ./cmd

./tempmail --new --fetch
./tempmail --count 5
./tempmail --email user@domain.com --fetch
./tempmail --domains
./tempmail --new --monitor 10
./tempmail --new --fetch --delete
```

## API Reference

The client has an "active account" (set by `CreateAccount`, `CreateAccountWithAddress`, `UseAccount` or `SetAccount`). The inbox helpers (`GetMessages`, `GetMessage`, `MarkMessageRead`, `DeleteMessage`, `WaitForMessage`, `Send`) operate on that account's `INBOX`. Everything else takes explicit IDs.

### Client options

```go
tempmail.NewClient(
    tempmail.WithAPIKey("smtplabs_..."),
    tempmail.WithAPIAddress("https://api.smtp.dev"),
    tempmail.WithTimeout(60*time.Second),
    tempmail.WithHTTPClient(customHTTPClient),
)
```

### Accounts

```go
account, err := client.CreateAccount("")                        // random address on an active domain
account, err := client.CreateAccountWithAddress("me@x.com", "") // password auto-generated if empty
account, err := client.UseAccount("me@x.com")                   // attach to an existing account
accounts, err := client.ListAccounts(1, "", nil)
account, err := client.GetAccountByID(id)
account, err := client.UpdateAccount(id, "newpass", &isActive)
err := client.DeleteAccount()                                   // active account
err := client.DeleteAccountByID(id)
```

### Domains

```go
names, err := client.GetDomains()                 // active domain names
domains, err := client.ListDomains(1, "", nil)
domain, err := client.CreateDomain("example.com", true)
domain, err := client.UpdateDomain(id, false)
err := client.DeleteDomain(id)
```

### Mailboxes

```go
mailboxes, err := client.ListMailboxes(accountID, 1, "")
mailbox, err := client.CreateMailbox(accountID, "Archive")
mailbox, err := client.UpdateMailbox(accountID, id, "Renamed")
err := client.DeleteMailbox(accountID, id)
```

### Messages (inbox helpers)

```go
messages, err := client.GetMessages(1)         // page number, newest first
msg, err := client.GetMessage("message-id")
err := client.MarkMessageRead("message-id")
err := client.DeleteMessage("message-id")
msg, err := client.WaitForMessage(ctx)         // blocks until a new message arrives
```

### Messages (any mailbox)

```go
messages, err := client.ListMessages(accountID, mailboxID, 1)
msg, err := client.FetchMessage(accountID, mailboxID, id)

read := true
msg, err := client.UpdateMessage(accountID, mailboxID, id, tempmail.MessageUpdate{IsRead: &read})

err := client.MoveMessage(accountID, mailboxID, id, targetMailboxID)
err := client.RemoveMessage(accountID, mailboxID, id)

raw, err := client.GetMessageSource(accountID, mailboxID, id)
eml, err := client.DownloadMessage(accountID, mailboxID, id)
data, err := client.DownloadAttachment(accountID, mailboxID, id, attachmentID)
```

### Sending

```go
err := client.Send(tempmail.SendMessageRequest{
    To:      []tempmail.Address{{Address: "you@example.com"}},
    Subject: "Hello",
    Text:    "Hi there",
})
```

### API tokens and Mercure

```go
token, err := client.CreateToken("ci", "used by CI") // token value is only returned once
tokens, err := client.ListTokens(1, "")
err := client.DeleteToken(id)

jwt, err := client.GetMercureToken()
```

For real-time updates, subscribe to `tempmail.MercureAddress` with `Authorization: Bearer <jwt>` and topic `/accounts/{id}{+path}`.

### Errors

Non-2xx responses return `*tempmail.APIError` (with `StatusCode` and `Message`). `errors.Is(err, tempmail.ErrNotFound)` matches 404s. The API is rate limited to 4096 requests/minute.

## Project Structure

```
.
├── cmd/main.go               # CLI application
├── example/main.go           # Example usage
├── pkg/tempmail/
│   ├── client.go             # Client, options, HTTP plumbing
│   ├── accounts.go
│   ├── domains.go
│   ├── mailboxes.go
│   ├── messages.go
│   ├── tokens.go             # API tokens + Mercure token
│   ├── errors.go
│   ├── models.go
│   └── utils.go
├── go.mod
└── README.md
```

## Running Tests

```bash
go test ./pkg/tempmail/... -v
```

## License

MIT
