package tempmail

import (
	"encoding/json"
	"sort"
)

type Domain struct {
	ID        string `json:"id"`
	Domain    string `json:"domain"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type Account struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	Password  string    `json:"password,omitempty"`
	Quota     int       `json:"quota"`
	Used      int       `json:"used"`
	IsActive  bool      `json:"isActive"`
	IsDeleted bool      `json:"isDeleted"`
	Mailboxes []Mailbox `json:"mailboxes"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

type Mailbox struct {
	ID                  string `json:"id"`
	Path                string `json:"path"`
	IsSystem            bool   `json:"isSystem"`
	TotalMessages       int    `json:"totalMessages"`
	TotalUnreadMessages int    `json:"totalUnreadMessages"`
	Account             string `json:"account,omitempty"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
}

type Address struct {
	Address string `json:"address"`
	Name    string `json:"name,omitempty"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Attachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Disposition string `json:"disposition"`
	CID         string `json:"cid,omitempty"`
	Size        int    `json:"size"`
	DownloadURL string `json:"downloadUrl"`
}

type HTMLBody []string

func (h *HTMLBody) UnmarshalJSON(data []byte) error {
	var list []string
	if err := json.Unmarshal(data, &list); err == nil {
		*h = list
		return nil
	}

	var byKey map[string]*string
	if err := json.Unmarshal(data, &byKey); err == nil {
		keys := make([]string, 0, len(byKey))
		for k := range byKey {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			if byKey[k] != nil {
				parts = append(parts, *byKey[k])
			}
		}
		*h = parts
		return nil
	}

	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*h = HTMLBody{single}
		return nil
	}

	*h = nil
	return nil
}

type Message struct {
	ID                string         `json:"id"`
	MsgID             string         `json:"msgid"`
	From              Address        `json:"from"`
	To                []Address      `json:"to"`
	CC                []Address      `json:"cc"`
	BCC               []Address      `json:"bcc"`
	ReplyTo           []Address      `json:"replyTo"`
	Date              string         `json:"date"`
	Subject           string         `json:"subject"`
	Intro             string         `json:"intro"`
	Text              string         `json:"text"`
	HTML              HTMLBody       `json:"html"`
	IsRead            bool           `json:"isRead"`
	IsFlagged         bool           `json:"isFlagged"`
	IsDeleted         bool           `json:"isDeleted"`
	HasAttachments    bool           `json:"hasAttachments"`
	Size              int            `json:"size"`
	AutoDeleteEnabled bool           `json:"autoDeleteEnabled"`
	ExpiresAt         string         `json:"expiresAt"`
	Flags             []string       `json:"flags"`
	Verifications     map[string]any `json:"verifications"`
	ThreadID          string         `json:"threadId"`
	Headers           []Header       `json:"headers"`
	Attachments       []Attachment   `json:"attachments"`
	DownloadURL       string         `json:"downloadUrl"`
	SourceURL         string         `json:"sourceUrl"`
	CreatedAt         string         `json:"createdAt"`
	UpdatedAt         string         `json:"updatedAt"`
}

type MessageUpdate struct {
	IsRead            *bool   `json:"isRead,omitempty"`
	IsFlagged         *bool   `json:"isFlagged,omitempty"`
	AutoDeleteEnabled *bool   `json:"autoDeleteEnabled,omitempty"`
	ExpiresAt         *string `json:"expiresAt,omitempty"`
}

type OutgoingAttachment struct {
	Content     string `json:"content"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType,omitempty"`
	Disposition string `json:"disposition,omitempty"`
	CID         string `json:"cid,omitempty"`
}

type SendMessageRequest struct {
	From        *Address             `json:"from,omitempty"`
	To          []Address            `json:"to"`
	CC          []Address            `json:"cc,omitempty"`
	BCC         []Address            `json:"bcc,omitempty"`
	ReplyTo     []Address            `json:"replyTo,omitempty"`
	Subject     string               `json:"subject,omitempty"`
	Text        string               `json:"text,omitempty"`
	HTML        string               `json:"html,omitempty"`
	Headers     []Header             `json:"headers,omitempty"`
	Attachments []OutgoingAttachment `json:"attachments,omitempty"`
}

type Token struct {
	ID          string `json:"id"`
	Token       string `json:"token,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type RawMessage struct {
	ID  string `json:"id"`
	Raw string `json:"raw"`
}

type collection[T any] struct {
	Member []T `json:"member"`
}
