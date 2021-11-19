package smartaccounts

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// SubscriptionSearchRequest represents the request required to return subscription data.
type SubscriptionSearchRequest struct {
	Source        string                                  `json:"source"`
	SmartAccounts []SubscriptionSearchRequestSmartAccount `json:"smartAccount"`
}

// SubscriptionSearchRequestSmartAccount represents the request details required to return subscription data.
type SubscriptionSearchRequestSmartAccount struct {
	SmartAccountID int    `json:"smartAccountId"`
	Domain         string `json:"domain"`
}

// SubscriptionSearchResponse represents the response from the Cisco API.
type SubscriptionSearchResponse struct {
	Source       string                           `json:"source"`
	Status       string                           `json:"status"`
	OfferDetails []SubscriptionSearchOfferDetails `json:"offerDetails"`
}

// SubscriptionSearchOfferDetails represents the offer details in the response from the Cisco API.
type SubscriptionSearchOfferDetails struct {
	SmartAccountID string                           `json:"smartAccountId"`
	Subscriptions  []SubscriptionSearchSubscription `json:"subscriptions"`
}

// SubscriptionSearchSubscription represents the subscription data in the offer details for the subscription search response.
type SubscriptionSearchSubscription struct {
	SubRefID              string                                   `json:"subRefId"`
	VirtualAccountDetails []SubscriptionSearchVirtualAccountDetail `json:"vaDetails"`
	Suites                []SubscriptionSearchSuite                `json:"suites"`
	AdditionalParameters  []SubscriptionSearchAdditionalParameters `json:"additionalParams"`
}

// SubscriptionSearchVirtualAccountDetail represents the virtual account detail for the subscription
type SubscriptionSearchVirtualAccountDetail struct {
	VirtualAccountID   string `json:"virtualAccountId"`
	VirtualAccountName string `json:"virtualAccountName"`
}

// SubscriptionSearchSuite represents the suites from the subscription search
type SubscriptionSearchSuite struct {
	SuiteName    string `json:"suiteName"`
	AtoName      string `json:"atoName"`
	Architecture string `json:"architecture"`
}

// SubscriptionSearchAdditionalParameters represents the additional parameters from the search
type SubscriptionSearchAdditionalParameters struct {
	ParameterName string `json:"paramName"`
	Value         string `json:"value"`
}

// SearchSubscriptions as per the Cisco documentation gets subscription details for BPA.  Given
// a smart account ID and domain it will search for subscriptions.  Note you may receive duplicates
// in the response since it is a search.
// https://apidocs-prod.cisco.com/explore;category=6083723a25042e9035f6a753;sgroup=6091ff087b37a601010bf23c;epname=614b1bc3b39ea324506c580d
func (c *Client) SearchSubscriptions(smartAccountID int, smartAccountDomain string) (*SubscriptionSearchResponse, error) {
	url := "https://swapi.cisco.com/services/api/smart-accounts-and-licensing/v1/subscription/search"
	payload, err := json.Marshal(&SubscriptionSearchRequest{
		Source:        "",
		SmartAccounts: []SubscriptionSearchRequestSmartAccount{{smartAccountID, smartAccountDomain}}})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	var ssr SubscriptionSearchResponse
	err = c.makeRequest(context.Background(), req, &ssr)
	if err != nil {
		return nil, err
	}
	return &ssr, nil

}
