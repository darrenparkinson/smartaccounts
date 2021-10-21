package smartaccounts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// Token represents a Cisco Access Token
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	ExpiresAt   time.Time
}

// Client represents the entry point to the library
type Client struct {
	clientID   string
	secret     string
	username   string
	password   string
	token      *Token
	lim        *rate.Limiter
	HTTPClient *http.Client
}

// Err implements the error interface so we can have constant errors.
type Err string

func (e Err) Error() string {
	return string(e)
}

var (
	ErrBadRequest    = Err("ccw: bad request")
	ErrUnauthorized  = Err("ccw: unauthorized request")
	ErrForbidden     = Err("ccw: forbidden")
	ErrNotFound      = Err("ccw: not found")
	ErrInternalError = Err("ccw: internal error")
	ErrUnknown       = Err("ccw: unexpected error occurred")
)

// SmartAccountResponse represents the top level response on requesting smart accounts
type SmartAccountResponse struct {
	Accounts      []SmartAccount `json:"accounts"`
	StatusMessage string         `json:"statusMessage"`
	Status        string         `json:"status"`
}

// SmartAccount represents an individual smart account, allowing you to easily add Virtual Accounts and Licenses
type SmartAccount struct {
	AccountStatus   string            `json:"accountStatus"`
	AccountDomain   string            `json:"accountDomain"`
	AccountName     string            `json:"accountName"`
	AccountType     string            `json:"accountType"`
	Roles           []Role            `json:"roles"`
	VirtualAccounts *[]VirtualAccount `json:"virtualAccounts"`
	Licenses        *[]License        `json:"licenses"`
}

// Role as specified by Cisco
type Role struct {
	Role string `json:"role"`
}

// VirtualAccountResponse represents the top level response from requesting virtual accounts for a domain.
type VirtualAccountResponse struct {
	VirtualAccounts []VirtualAccount `json:"virtualAccounts"`
	StatusMessage   string           `json:"statusMessage"`
	Status          string           `json:"status"`
}

// VirtualAccount represents an individual virtual account
type VirtualAccount struct {
	IsDefault           string `json:"isDefault"` // Really a bool in quotes. TODO: Add custom unmarshal
	Name                string `json:"name"`
	Description         string `json:"description"`
	CommerceAccessLevel string `json:"commerceAccessLevel"`
}

// SearchResponse represents the top level response for a search
type SearchResponse struct {
	TotalRecords  int             `json:"totalRecords"`
	Accounts      []SearchAccount `json:"accounts"`
	StatusMessage string          `json:"statusMessage"`
	Status        string          `json:"status"`
}

// SearchAccount represents the detail returned from a Search which is not the same as a SmartAccount unfortunately
type SearchAccount struct {
	Domain string `json:"domain"`
	Name   string `json:"name"`
	ID     int    `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// LicenseRequest represents the details required to fetch license usage details
type LicenseRequest struct {
	VirtualAccounts []string `json:"virtualAccounts"`
	Limit           int      `json:"limit"`
	Offset          int      `json:"offset"`
}

// LicenseResponse represents the top level response for a request for licenses
type LicenseResponse struct {
	TotalRecords  int       `json:"totalRecords"`
	Licenses      []License `json:"licenses"`
	StatusMessage string    `json:"statusMessage"`
	Status        string    `json:"status"`
}

// License represents an individual license
type License struct {
	LicenseSubstitutions []LicenseSubstitution `json:"licenseSubstitutions"`
	IsPortable           bool                  `json:"isPortable"`
	License              string                `json:"license"`
	VirtualAccount       string                `json:"virtualAccount"`
	Quantity             int                   `json:"quantity"`
	InUse                int                   `json:"inUse"`
	Available            int                   `json:"available"`
	Status               string                `json:"status"`
	BillingType          string                `json:"billingType"` // PREPAID or USAGE
	AhaApps              bool                  `json:"ahaApps"`
	PendingQuantity      int                   `json:"pendingQuantity"`
	Reserved             int                   `json:"reserved"`
	LicenseDetails       []LicenseDetail       `json:"licenseDetails"`
}

// LicenseSubstitution represents a license substitution detail from a license
type LicenseSubstitution struct {
	LicenseName         string `json:"licenseName"`
	SubstitutedLicense  string `json:"substitutedLicense"`
	SubstitutedQuantity int    `json:"substitutedQuantity"`
	SubstitutionType    string `json:"substitutionType"`
}

// LicenseDetail represents the license detail object from a license
type LicenseDetail struct {
	LicenseType    string `json:"licenseType"` // TERM/DEMO/PERPETUAL
	Quantity       int    `json:"quantity"`
	StartDate      string `json:"startDate"`
	EndDate        string `json:"endDate"`
	SubscriptionID string `json:"subscriptionId"`
	Status         string `json:"status"`
}

// New returns a new CCW client for accessing the smart accounts API
func New(client_id, client_secret, username, password string) *Client {
	limiter := rate.NewLimiter(100, 1)
	return &Client{
		clientID: client_id,
		secret:   client_secret,
		username: username,
		password: password,
		lim:      limiter,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetSmartLicenseUsage returns the Smart License Usage as per the Cisco documentation:
// https://apidocs-prod.cisco.com/explore;category=6083723a25042e9035f6a753;sgroup=6083723b25042e9035f6a775;epname=6131c97117b4092245f49d9f
// Requires the provided SmartAccount to have the AccountDomain field specified and a list of virtual accounts populated.
func (c *Client) GetSmartLicenseUsage(sa SmartAccount) (*[]License, error) {
	licenses := []License{}
	for _, va := range *sa.VirtualAccounts {
		// log.Println("retrieving licenses for", sa.AccountDomain, va.Name)
		offset, limit := 0, 100
		for {
			url := fmt.Sprintf("https://apx.cisco.com/services/api/smart-accounts-and-licensing/v1/accounts/%s/licenses", sa.AccountDomain)
			method := "POST"
			payload, err := json.Marshal(&LicenseRequest{Offset: offset, Limit: limit, VirtualAccounts: []string{va.Name}})
			if err != nil {
				return nil, err
			}
			req, err := http.NewRequest(method, url, bytes.NewReader(payload))
			if err != nil {
				return nil, err
			}
			var lr LicenseResponse
			err = c.makeRequest(context.Background(), req, &lr)
			if err != nil {
				log.Printf("error retrieving licenses for %s: %s: %s", sa.AccountDomain, va.Name, err)
				break
			}
			licenses = append(licenses, lr.Licenses...)
			if lr.TotalRecords < limit {
				break
			}
			offset += limit
			if offset > lr.TotalRecords {
				break
			}
		}
	}
	return &licenses, nil
}

// SearchSmartAccountsByDomain will return any entry that matches your search, so be careful, since a search for
// e.g. work.com will return wework.com, wewontwork.com, wedontwork.com etc.
// Also note that there is a hardcoded limit of 1000 entries for the response.
func (c *Client) SearchSmartAccountsByDomain(domain string) (*SearchResponse, error) {
	url := fmt.Sprintf("https://apx.cisco.com/services/api/smart-accounts-and-licensing/v1/accounts/search?domain=%s&type=CUSTOMER&limit=1000&offset=0", domain)
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	var sr SearchResponse
	err = c.makeRequest(context.Background(), req, &sr)
	if err != nil {
		return nil, err
	}
	return &sr, nil
}

// GetVirtualAccounts will retrieve a list of virtual accounts given a valid smart account domain.
func (c *Client) GetVirtualAccounts(domain string) ([]VirtualAccount, error) {
	url := fmt.Sprintf("https://swapi.cisco.com/services/api/smart-accounts-and-licensing/v1/accounts/%s/customer/virtual-accounts", domain)
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	var varesp VirtualAccountResponse
	err = c.makeRequest(context.Background(), req, &varesp)
	if err != nil {
		return nil, err
	}
	return varesp.VirtualAccounts, nil
}

// GetAllSmartAccounts will retrieve a list of all smart accounts the user account has access to.  Note that
// this does not (rather annoyingly) return the Smart Account ID that you will likely need.  For that you
// will have to use SearchSmartAccountsByDomain and match them up yourself.
func (c *Client) GetAllSmartAccounts() ([]SmartAccount, error) {
	url := "https://swapi.cisco.com/services/api/smart-accounts-and-licensing/v2/accounts"
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	var sar SmartAccountResponse
	err = c.makeRequest(context.Background(), req, &sar)
	if err != nil {
		return nil, err
	}
	return sar.Accounts, nil
}

// makeRequest provides a single function to add common items to the request.
func (c *Client) makeRequest(ctx context.Context, req *http.Request, v interface{}) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	if !c.lim.Allow() {
		c.lim.Wait(ctx)
	}

	rc := req.WithContext(ctx)
	res, err := c.HTTPClient.Do(rc)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	// if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
	if res.StatusCode != http.StatusOK {
		var ccwErr error
		switch res.StatusCode {
		case 400:
			ccwErr = ErrBadRequest
		case 401:
			ccwErr = ErrUnauthorized
		case 403:
			ccwErr = ErrForbidden
		case 404:
			ccwErr = ErrNotFound
		case 500:
			ccwErr = ErrInternalError
		default:
			// ccwErr = ErrUnknown
			ccwErr = fmt.Errorf("unknown error: %s", res.Status)
		}
		return ccwErr
	}
	if res.StatusCode == http.StatusCreated {
		return nil
	}
	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return err
	}
	return nil
}

// getToken returns a new token for use with the SmartAccounts API.  It can be used as required since
// it will memoise an existing token until 5 minutes before expiry.
func (c *Client) getToken() (*Token, error) {
	now := time.Now().UTC()
	if c.token != nil && c.token.ExpiresAt.Sub(now).Minutes() > 5 {
		return c.token, nil
	}
	url := "https://cloudsso.cisco.com/as/token.oauth2"
	method := "POST"
	pl := fmt.Sprintf("client_id=%s&client_secret=%s&username=%s&password=%s&grant_type=password", c.clientID, c.secret, c.username, c.password)
	payload := strings.NewReader(pl)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var t Token
	err = json.NewDecoder(res.Body).Decode(&t)
	if err != nil {
		return nil, err
	}
	t.ExpiresAt = time.Unix(now.Unix()+t.ExpiresIn, 0)
	c.token = &t
	return &t, nil
}
