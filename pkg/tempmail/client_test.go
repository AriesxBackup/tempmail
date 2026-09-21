package tempmail

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-API-KEY"); got != "test-key" {
			t.Errorf("expected X-API-KEY test-key, got %q", got)
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	return NewClient(WithAPIAddress(server.URL), WithAPIKey("test-key"))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/ld+json")
	json.NewEncoder(w).Encode(v)
}

func withInbox(c *Client) *Client {
	c.account = &Account{
		ID:        "acc1",
		Address:   "test@test.com",
		Mailboxes: []Mailbox{{ID: "mb1", Path: "INBOX"}},
	}
	return c
}

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client.apiAddress != APIAddress {
		t.Errorf("expected apiAddress %s, got %s", APIAddress, client.apiAddress)
	}

	if client.HasAPIKey() {
		t.Error("new client should not have an API key")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	client := NewClient(WithAPIAddress("https://custom.api.com/"), WithAPIKey("k"), WithTimeout(time.Second))

	if client.apiAddress != "https://custom.api.com" {
		t.Errorf("unexpected apiAddress %s", client.apiAddress)
	}

	if !client.HasAPIKey() {
		t.Error("expected API key to be set")
	}

	if client.httpClient.Timeout != time.Second {
		t.Errorf("unexpected timeout %v", client.httpClient.Timeout)
	}
}

func TestRequiresAPIKey(t *testing.T) {
	client := NewClient()

	if _, err := client.GetDomains(); !errors.Is(err, ErrAPIKeyRequired) {
		t.Errorf("expected ErrAPIKeyRequired, got %v", err)
	}
}

func TestGetDomains(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/domains" {
			t.Errorf("expected path /domains, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("isActive") != "true" {
			t.Errorf("expected isActive=true, got %q", r.URL.Query().Get("isActive"))
		}

		writeJSON(w, collection[Domain]{Member: []Domain{
			{ID: "1", Domain: "test.com", IsActive: true},
			{ID: "2", Domain: "off.com", IsActive: false},
			{ID: "3", Domain: "example.com", IsActive: true},
		}})
	})

	domains, err := client.GetDomains()
	if err != nil {
		t.Fatalf("GetDomains failed: %v", err)
	}

	if len(domains) != 2 || domains[0] != "test.com" || domains[1] != "example.com" {
		t.Errorf("unexpected domains: %v", domains)
	}
}

func TestGetDomainsEmpty(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, collection[Domain]{})
	})

	if _, err := client.GetDomains(); !errors.Is(err, ErrNoDomains) {
		t.Errorf("expected ErrNoDomains, got %v", err)
	}
}

func TestCreateAccount(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/domains":
			writeJSON(w, collection[Domain]{Member: []Domain{{ID: "1", Domain: "test.com", IsActive: true}}})
		case r.URL.Path == "/accounts" && r.Method == http.MethodPost:
			if ct := r.Header.Get("Content-Type"); ct != "application/ld+json" {
				t.Errorf("unexpected content type %s", ct)
			}

			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["password"] == "" || body["address"] == "" {
				t.Errorf("missing address/password: %v", body)
			}

			w.WriteHeader(http.StatusCreated)
			writeJSON(w, Account{
				ID:        "acc1",
				Address:   body["address"].(string),
				IsActive:  true,
				Mailboxes: []Mailbox{{ID: "mb1", Path: "INBOX"}},
			})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	account, err := client.CreateAccount("")
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	if account.ID != "acc1" || account.Password == "" {
		t.Errorf("unexpected account: %+v", account)
	}

	if client.GetAccount() != account {
		t.Error("client should track the created account")
	}
}

func TestUseAccount(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("address") == "" {
			t.Errorf("expected address query, got %q", r.URL.RawQuery)
		}
		writeJSON(w, collection[Account]{Member: []Account{{ID: "acc9", Address: "me@test.com"}}})
	})

	account, err := client.UseAccount("me@test.com")
	if err != nil {
		t.Fatalf("UseAccount failed: %v", err)
	}
	if account.ID != "acc9" {
		t.Errorf("unexpected account %+v", account)
	}

	if _, err := client.UseAccount("other@test.com"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetMessages(t *testing.T) {
	client := withInbox(newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/accounts/acc1/mailboxes/mb1/messages" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("expected page=2, got %q", r.URL.Query().Get("page"))
		}

		w.Write([]byte(`{"member":[
			{"id":"m1","subject":"Hi","from":{"address":"a@b.com","name":null},"text":"body","html":{"0":"<p>x</p>"},"isRead":false,"expiresAt":null},
			{"id":"m2","subject":"Yo","html":["<b>y</b>"],"isRead":true}
		]}`))
	}))

	messages, err := client.GetMessages(2)
	if err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}

	if messages[0].From.Address != "a@b.com" || messages[0].Text != "body" {
		t.Errorf("unexpected message: %+v", messages[0])
	}

	if len(messages[0].HTML) != 1 || messages[0].HTML[0] != "<p>x</p>" {
		t.Errorf("unexpected html: %v", messages[0].HTML)
	}

	if len(messages[1].HTML) != 1 || !messages[1].IsRead {
		t.Errorf("unexpected second message: %+v", messages[1])
	}
}

func TestGetMessagesNoAccount(t *testing.T) {
	client := NewClient(WithAPIKey("k"))

	if _, err := client.GetMessages(1); !errors.Is(err, ErrNoActiveAccount) {
		t.Errorf("expected ErrNoActiveAccount, got %v", err)
	}
}

func TestInboxLookupFallback(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/accounts/acc1/mailboxes":
			if r.URL.Query().Get("path") != "INBOX" {
				t.Errorf("expected path=INBOX, got %q", r.URL.RawQuery)
			}
			writeJSON(w, collection[Mailbox]{Member: []Mailbox{{ID: "mbX", Path: "INBOX"}}})
		case "/accounts/acc1/mailboxes/mbX/messages":
			writeJSON(w, collection[Message]{})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})
	client.account = &Account{ID: "acc1"}

	if _, err := client.GetMessages(1); err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}
}

func TestGetMessage(t *testing.T) {
	client := withInbox(newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/accounts/acc1/mailboxes/mb1/messages/msg123" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		writeJSON(w, Message{ID: "msg123", Subject: "Test", Text: "Hello World"})
	}))

	msg, err := client.GetMessage("msg123")
	if err != nil {
		t.Fatalf("GetMessage failed: %v", err)
	}

	if msg.ID != "msg123" || msg.Text != "Hello World" {
		t.Errorf("unexpected message: %+v", msg)
	}
}

func TestMarkMessageRead(t *testing.T) {
	client := withInbox(newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/merge-patch+json" {
			t.Errorf("unexpected content type %s", ct)
		}

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["isRead"] != true {
			t.Errorf("expected isRead true, got %v", body)
		}

		writeJSON(w, Message{ID: "msg123", IsRead: true})
	}))

	if err := client.MarkMessageRead("msg123"); err != nil {
		t.Fatalf("MarkMessageRead failed: %v", err)
	}
}

func TestDeleteMessage(t *testing.T) {
	client := withInbox(newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := client.DeleteMessage("msg123"); err != nil {
		t.Fatalf("DeleteMessage failed: %v", err)
	}
}

func TestMoveMessage(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/accounts/acc1/mailboxes/mb1/messages/m1/move" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["mailbox"] != "/accounts/acc1/mailboxes/mb2" {
			t.Errorf("unexpected body %v", body)
		}

		writeJSON(w, Message{ID: "m1"})
	})

	if err := client.MoveMessage("acc1", "mb1", "m1", "mb2"); err != nil {
		t.Fatalf("MoveMessage failed: %v", err)
	}
}

func TestSendMessage(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/accounts/acc1/messages/send" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}

		var body SendMessageRequest
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.To) != 1 || body.To[0].Address != "to@test.com" || body.Text != "hi" {
			t.Errorf("unexpected body %+v", body)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	client.account = &Account{ID: "acc1"}

	err := client.Send(SendMessageRequest{To: []Address{{Address: "to@test.com"}}, Subject: "s", Text: "hi"})
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
}

func TestDownloadAttachment(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/accounts/acc1/mailboxes/mb1/messages/m1/attachment/att1" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Write([]byte("filedata"))
	})

	data, err := client.DownloadAttachment("acc1", "mb1", "m1", "att1")
	if err != nil {
		t.Fatalf("DownloadAttachment failed: %v", err)
	}
	if string(data) != "filedata" {
		t.Errorf("unexpected data %q", data)
	}
}

func TestDeleteAccount(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/accounts/acc1" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	client.account = &Account{ID: "acc1", Address: "test@test.com"}

	if err := client.DeleteAccount(); err != nil {
		t.Fatalf("DeleteAccount failed: %v", err)
	}

	if client.account != nil {
		t.Error("account should be nil after deletion")
	}

	if err := client.DeleteAccount(); !errors.Is(err, ErrNoActiveAccount) {
		t.Errorf("expected ErrNoActiveAccount, got %v", err)
	}
}

func TestMercureToken(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mercure/token" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Write([]byte(`{"token":"eyJ.abc"}`))
	})

	token, err := client.GetMercureToken()
	if err != nil || token != "eyJ.abc" {
		t.Errorf("unexpected result %q, %v", token, err)
	}
}

func TestAPIError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"detail":"Not Found"}`))
	})

	_, err := client.GetDomain("nope")

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 || apiErr.Message != "Not Found" {
		t.Fatalf("unexpected error %v", err)
	}

	if !errors.Is(err, ErrNotFound) {
		t.Error("404 should match ErrNotFound")
	}
}

func TestWaitForMessage(t *testing.T) {
	calls := 0
	client := withInbox(newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			writeJSON(w, collection[Message]{Member: []Message{{ID: "old"}}})
			return
		}
		writeJSON(w, collection[Message]{Member: []Message{{ID: "new"}, {ID: "old"}}})
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	msg, err := client.WaitForMessage(ctx)
	if err != nil {
		t.Fatalf("WaitForMessage failed: %v", err)
	}
	if msg.ID != "new" {
		t.Errorf("expected new message, got %s", msg.ID)
	}
}

func TestWaitForMessageTimeout(t *testing.T) {
	client := withInbox(newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, collection[Message]{})
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if _, err := client.WaitForMessage(ctx); !errors.Is(err, ErrTimeout) {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
}
