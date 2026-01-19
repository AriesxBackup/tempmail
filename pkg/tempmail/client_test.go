package tempmail

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.apiAddress != APIAddress {
		t.Errorf("expected apiAddress %s, got %s", APIAddress, client.apiAddress)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	customURL := "https://custom.api.com"
	client := NewClient(WithAPIAddress(customURL))

	if client.apiAddress != customURL {
		t.Errorf("expected apiAddress %s, got %s", customURL, client.apiAddress)
	}
}

func TestGetDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/domains" {
			t.Errorf("expected path /domains, got %s", r.URL.Path)
		}

		resp := DomainResponse{
			Member: []Domain{
				{ID: "1", Domain: "test.com"},
				{ID: "2", Domain: "example.com"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithAPIAddress(server.URL))
	domains, err := client.GetDomains()

	if err != nil {
		t.Fatalf("GetDomains failed: %v", err)
	}

	if len(domains) != 2 {
		t.Errorf("expected 2 domains, got %d", len(domains))
	}

	if domains[0] != "test.com" {
		t.Errorf("expected first domain test.com, got %s", domains[0])
	}
}

func TestLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			t.Errorf("expected path /token, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		resp := TokenResponse{
			Token: "test-token-123",
			ID:    "acc123",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithAPIAddress(server.URL))
	err := client.Login("test@test.com", "password123")

	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if !client.IsAuthenticated() {
		t.Error("client should be authenticated")
	}
}

func TestGetMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/messages/msg123" {
			t.Errorf("expected path /messages/msg123, got %s", r.URL.Path)
		}

		resp := Message{
			ID:      "msg123",
			Subject: "Test Subject",
			Text:    "Hello World",
			From:    Address{Address: "sender@test.com"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithAPIAddress(server.URL))
	client.token = "test-token"
	client.authHeaders = map[string]string{
		"Authorization": "Bearer test-token",
	}

	msg, err := client.GetMessage("msg123")
	if err != nil {
		t.Fatalf("GetMessage failed: %v", err)
	}

	if msg.ID != "msg123" {
		t.Errorf("expected ID msg123, got %s", msg.ID)
	}

	if msg.Text != "Hello World" {
		t.Errorf("expected text 'Hello World', got '%s'", msg.Text)
	}
}

func TestGetMessageNotAuthenticated(t *testing.T) {
	client := NewClient()
	_, err := client.GetMessage("msg123")

	if err != ErrNotAuthenticated {
		t.Errorf("expected ErrNotAuthenticated, got %v", err)
	}
}

func TestDeleteMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(WithAPIAddress(server.URL))
	client.token = "test-token"
	client.authHeaders = map[string]string{
		"Authorization": "Bearer test-token",
	}

	err := client.DeleteMessage("msg123")
	if err != nil {
		t.Fatalf("DeleteMessage failed: %v", err)
	}
}

func TestDeleteAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(WithAPIAddress(server.URL))
	client.token = "test-token"
	client.authHeaders = map[string]string{
		"Authorization": "Bearer test-token",
	}
	client.account = &Account{ID: "acc123", Address: "test@test.com"}

	err := client.DeleteAccount()
	if err != nil {
		t.Fatalf("DeleteAccount failed: %v", err)
	}

	if client.account != nil {
		t.Error("account should be nil after deletion")
	}

	if client.token != "" {
		t.Error("token should be empty after deletion")
	}
}

func TestMarkMessageSeen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/merge-patch+json" {
			t.Errorf("expected Content-Type application/merge-patch+json, got %s", contentType)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"seen": true})
	}))
	defer server.Close()

	client := NewClient(WithAPIAddress(server.URL))
	client.token = "test-token"

	err := client.MarkMessageSeen("msg123")
	if err != nil {
		t.Fatalf("MarkMessageSeen failed: %v", err)
	}
}

func TestIsAuthenticated(t *testing.T) {
	client := NewClient()

	if client.IsAuthenticated() {
		t.Error("new client should not be authenticated")
	}

	client.token = "test-token"

	if !client.IsAuthenticated() {
		t.Error("client with token should be authenticated")
	}
}

func TestGetAccount(t *testing.T) {
	client := NewClient()

	if client.GetAccount() != nil {
		t.Error("new client should have nil account")
	}

	client.account = &Account{ID: "123", Address: "test@test.com"}

	acc := client.GetAccount()
	if acc == nil {
		t.Error("account should not be nil")
	}

	if acc.ID != "123" {
		t.Errorf("expected ID 123, got %s", acc.ID)
	}
}
