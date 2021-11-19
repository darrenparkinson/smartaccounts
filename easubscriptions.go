package smartaccounts

import (
	"context"
	"fmt"
	"net/http"
)

// EAConsumptionReportError represents the error received by GetEASmartAccountSubscriptionConsumptionReport which
// Cisco sends as a 400 Bad Request, typically when there are no subscriptions for the provided details.
type EAConsumptionReportError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// EASmartAccountSubscriptionConsumptionReportResponse represents the response from the EA Smart Account
// Subscription Consumption Report.  Sorry for the ridiculously long name!
type EASmartAccountSubscriptionConsumptionReportResponse struct {
	Subscriptions []EASubscription `json:"subscriptions"`
}

// EASubscription represents a subscription from the EA Consumption Report
type EASubscription struct {
	SubscriptionID            string      `json:"subscriptionID"`
	Status                    string      `json:"status"`
	StartDate                 string      `json:"startDate"`
	EndDate                   string      `json:"endDate"`
	Duration                  int         `json:"duration"`
	RemainingDuration         int         `json:"remainingDuration"`
	DurationInMonths          int         `json:"durationInMonths"`
	RemainingDurationInMonths int         `json:"remainingDurationInMonths"`
	NextTrueForward           string      `json:"nextTrueForward"`
	ArchitectureName          string      `json:"architectureName"`
	Accounts                  []EAAccount `json:"accounts"`
}

// EAAccount represents the Account from the EA Consumption Report Subscription
type EAAccount struct {
	SmartAccountID   int                `json:"smartAccountId"`
	SmartAccountName string             `json:"smartAccountName"`
	VirtualAccounts  []EAVirtualAccount `json:"vitualAccounts"` // NOTE THE TYPO!!!
}

// EAVirtualAccount represents the Virtual Account from the EA Consumption Report Subscription Account
type EAVirtualAccount struct {
	VirtualAccountID   int       `json:"virtualAccountId"`
	VirtualAccountName string    `json:"virtualAccountName"`
	Suites             []EASuite `json:"suites"`
}

// EASuite represents the Suite from the EA Consumption Report Subscription Virtual Account
type EASuite struct {
	CustSuiteID           int             `json:"custSuiteId"`
	SuiteName             string          `json:"suiteName"`
	CustSuiteName         string          `json:"custSuiteName"`
	PurchasedEntitlements int             `json:"purchasedEntitlements"`
	PremierEntitlements   int             `json:"premierEntitlements"`
	GrowthAllowance       int             `json:"growthAllowance"`
	TotalEntitlements     int             `json:"totalEntitlements"`
	PreEAConsumption      int             `json:"preEAConsumption"`
	LicenseGenerated      int             `json:"licenseGenerated"`
	LicenseMigrated       int             `json:"licenseMigrated"`
	C1ToDNAMigratedCount  int             `json:"c1ToDNAMigratedCount"`
	TotalConsumption      int             `json:"totalConsumption"`
	RemainingEntitlements int             `json:"remainingEntitlements"`
	SoftwareDownloads     int             `json:"softwareDownloads"`
	HealthMessage         string          `json:"healthMessage"`
	CalculationMethod     string          `json:"calculationMethod"`
	CommitmentType        string          `json:"commitmentType"`
	CommerceSkUs          []EACommerceSKU `json:"commerceSkus"`
}

// EACommerceSKU represents the actual line item from the EA Consumption Report Subscription Suite
type EACommerceSKU struct {
	EOL                    bool   `json:"eol"`
	CustSuiteID            int    `json:"custSuiteId"`
	CommerceSKU            string `json:"commerceSku"`
	CommerceSKUDescription string `json:"commerceSkuDescription"`
	SuiteName              string `json:"suiteName"`
	CustSuiteName          string `json:"custSuiteName"`
	EOLMessage             string `json:"eolMessage"`
	PurchasedEntitlements  int    `json:"purchasedEntitlements"`
	PremierEntitlements    int    `json:"premierEntitlements"`
	GrowthAllowance        int    `json:"growthAllowance"`
	TotalEntitlements      int    `json:"totalEntitlements"`
	PreEAConsumption       int    `json:"preEAConsumption"`
	LicenseGenerated       int    `json:"licenseGenerated"`
	LicenseMigrated        int    `json:"licenseMigrated"`
	C1ToDNAMigratedCount   int    `json:"c1ToDNAMigratedCount"`
	TotalConsumption       int    `json:"totalConsumption"`
	RemainingEntitlements  int    `json:"remainingEntitlements"`
	SoftwareDownloads      int    `json:"softwareDownloads"`
	HealthMessage          string `json:"healthMessage,omitempty"`
	CalculationMethod      string `json:"calculationMethod"`
	CommitmentType         string `json:"commitmentType"`
}

// GetEASmartAccountSubscriptionConsumptionReport can be used to get the consumption report for the EA
// Subscriptions.
func (c *Client) GetEASmartAccountSubscriptionConsumptionReport(smartAccountDomain, subscriptionID string) (*EASmartAccountSubscriptionConsumptionReportResponse, error) {
	url := fmt.Sprintf("https://swapi.cisco.com/services/api/enterprise-agreements/v1/subscription/account/%s/subscription/%s/consumption", smartAccountDomain, subscriptionID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var ear EASmartAccountSubscriptionConsumptionReportResponse
	err = c.makeRequest(context.Background(), req, &ear)
	if err != nil {
		return nil, err
	}
	return &ear, nil
}
