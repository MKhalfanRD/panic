package midtrans

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	serverKey string
	clientKey string
	isProd    bool
	http      *http.Client
}

type SnapInput struct {
	OrderID     string
	GrossAmount int64
	Name        string
	Email       string
	Phone       string
	ItemID      string
	ItemName    string
	FinishURL   string
}

type SnapResult struct {
	Token       string
	RedirectURL string
}

type Notification struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
}

func NewClient(serverKey, clientKey string, isProd bool) *Client {
	return &Client{
		serverKey: serverKey,
		clientKey: clientKey,
		isProd:    isProd,
		http:      &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *Client) Configured() bool {
	return c != nil && strings.TrimSpace(c.serverKey) != ""
}

func (c *Client) ClientKey() string {
	if c == nil {
		return ""
	}
	return c.clientKey
}

func (c *Client) SnapJSURL() string {
	if c != nil && c.isProd {
		return "https://app.midtrans.com/snap/snap.js"
	}
	return "https://app.sandbox.midtrans.com/snap/snap.js"
}

func (c *Client) CreateSnap(in SnapInput) (SnapResult, error) {
	if !c.Configured() {
		return SnapResult{}, fmt.Errorf("midtrans belum diatur")
	}
	payload := map[string]any{
		"transaction_details": map[string]any{
			"order_id":     in.OrderID,
			"gross_amount": in.GrossAmount,
		},
		"customer_details": map[string]string{
			"first_name": clip(in.Name, 40),
			"email":      in.Email,
			"phone":      in.Phone,
		},
		"item_details": []map[string]any{{
			"id":       clip(in.ItemID, 30),
			"price":    in.GrossAmount,
			"quantity": 1,
			"name":     clip(in.ItemName, 50),
		}},
	}
	if strings.TrimSpace(in.FinishURL) != "" {
		payload["callbacks"] = map[string]string{"finish": in.FinishURL}
	}
	var out struct {
		Token       string `json:"token"`
		RedirectURL string `json:"redirect_url"`
	}
	if err := c.post(c.snapURL(), payload, &out); err != nil {
		return SnapResult{}, err
	}
	if strings.TrimSpace(out.Token) == "" {
		return SnapResult{}, fmt.Errorf("midtrans snap tidak mengembalikan token")
	}
	return SnapResult{Token: out.Token, RedirectURL: out.RedirectURL}, nil
}

func (c *Client) Status(orderID string) (Notification, error) {
	var note Notification
	if !c.Configured() {
		return note, fmt.Errorf("midtrans belum diatur")
	}
	endpoint := c.apiBase() + "/v2/" + url.PathEscape(orderID) + "/status"
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return note, err
	}
	c.authorize(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return note, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return note, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return note, fmt.Errorf("midtrans status %d", resp.StatusCode)
	}
	if err := json.Unmarshal(raw, &note); err != nil {
		return note, err
	}
	return note, nil
}

func (c *Client) post(endpoint string, payload any, dest any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	c.authorize(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("midtrans snap %d", resp.StatusCode)
	}
	return json.Unmarshal(raw, dest)
}

func (c *Client) authorize(req *http.Request) {
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.serverKey+":")))
}

func (c *Client) apiBase() string {
	if c.isProd {
		return "https://api.midtrans.com"
	}
	return "https://api.sandbox.midtrans.com"
}

func (c *Client) snapURL() string {
	if c.isProd {
		return "https://app.midtrans.com/snap/v1/transactions"
	}
	return "https://app.sandbox.midtrans.com/snap/v1/transactions"
}

func clip(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n])
}
