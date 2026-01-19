package tempmail

type Account struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	Password string `json:"password,omitempty"`
}

type Domain struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
}

type Address struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type Message struct {
	ID        string    `json:"id"`
	From      Address   `json:"from"`
	To        []Address `json:"to"`
	Subject   string    `json:"subject"`
	Intro     string    `json:"intro"`
	Text      string    `json:"text"`
	HTML      []string  `json:"html"`
	Seen      bool      `json:"seen"`
	CreatedAt string    `json:"createdAt"`
}

type HydraResponse struct {
	Member     []Message `json:"hydra:member"`
	TotalItems int       `json:"hydra:totalItems"`
}

type DomainResponse struct {
	Member []Domain `json:"hydra:member"`
}

type TokenResponse struct {
	Token string `json:"token"`
	ID    string `json:"id"`
}
