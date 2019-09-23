package notionapi

type SubscriptionDataSpaceUsers struct {
	UserID       string        `json:"userId"`
	Role         string        `json:"role"`
	IsGuest      bool          `json:"isGuest"`
	GuestPageIds []interface{} `json:"guestPageIds"`
}

type SubscriptionDataCredits struct {
	ID               string `json:"id"`
	Version          int    `json:"version"`
	UserID           string `json:"user_id"`
	Amount           int    `json:"amount"`
	Activated        bool   `json:"activated"`
	CreatedTimestamp string `json:"created_timestamp"`
	Type             string `json:"type"`
}

type SubscriptionDataAddress struct {
	Name         string `json:"name"`
	BusinessName string `json:"businessName"`
	AddressLine1 string `json:"addressLine1"`
	AddressLine2 string `json:"addressLine2"`
	ZipCode      string `json:"zipCode"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
}

type SubscriptionData struct {
	Type              string                       `json:"type"`
	SpaceUsers        []SubscriptionDataSpaceUsers `json:"spaceUsers"`
	Credits           []SubscriptionDataCredits    `json:"credits"`
	TotalCredit       int                          `json:"totalCredit"`
	AvailableCredit   int                          `json:"availableCredit"`
	CreditEnabled     bool                         `json:"creditEnabled"`
	CustomerID        string                       `json:"customerId"`
	CustomerName      string                       `json:"customerName"`
	VatID             string                       `json:"vatId"`
	IsDelinquent      bool                         `json:"isDelinquent"`
	ProductID         string                       `json:"productId"`
	BillingEmail      string                       `json:"billingEmail"`
	Plan              string                       `json:"plan"`
	PlanAmount        int                          `json:"planAmount"`
	AccountBalance    int                          `json:"accountBalance"`
	MonthlyPlanAmount int                          `json:"monthlyPlanAmount"`
	YearlyPlanAmount  int                          `json:"yearlyPlanAmount"`
	Quantity          int                          `json:"quantity"`
	Billing           string                       `json:"billing"`
	Address           SubscriptionDataAddress      `json:"address"`
	Last4             string                       `json:"last4"`
	Brand             string                       `json:"brand"`
	Interval          string                       `json:"interval"`
	Created           int64                        `json:"created"`
	PeriodEnd         int64                        `json:"periodEnd"`
	NextInvoiceTime   int64                        `json:"nextInvoiceTime"`
	NextInvoiceAmount int                          `json:"nextInvoiceAmount"`
	IsPaid            bool                         `json:"isPaid"`
	Members           []interface{}                `json:"members"`

	RawJSON map[string]interface{} `json:"-"`
}

// GetSubscriptionData executes a raw API call /api/v3/getSubscriptionData
func (c *Client) GetSubscriptionData(spaceID string) (*SubscriptionData, error) {
	req := &struct {
		SpaceID string `json:"spaceId"`
	}{
		SpaceID: spaceID,
	}

	apiURL := "/api/v3/getSubscriptionData"
	var rsp SubscriptionData
	var err error
	rsp.RawJSON, err = doNotionAPI(c, apiURL, req, &rsp)
	if err != nil {
		return nil, err
	}

	return &rsp, nil
}
