package nordigen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Transactions contains booked and pending transactions.
type Transactions struct {
	Booked  []Transaction `json:"booked"`
	Pending []Transaction `json:"pending"`
}

// Transaction represents a transaction.
type Transaction struct {
	TransactionID string `json:"transactionId"`

	TransactionAmount Amount              `json:"transactionAmount"`
	CurrencyExchange  []*CurrencyExchange `json:"currencyExchange"`

	BankTransactionCode                    string   `json:"bankTransactionCode"`
	AdditionalInformation                  string   `json:"additionalInformation"`
	RemittanceInformationUnstructuredArray []string `json:"remittanceInformationUnstructuredArray"`
	BookingDate                            Date     `json:"bookingDate"`
	ValueDate                              Date     `json:"valueDate"`

	DebtorName    string       `json:"debtorName"`
	DebtorAccount *IBANAccount `json:"debtorAccount"`

	CreditorName    string       `json:"creditorName"`
	CreditorAccount *IBANAccount `json:"creditorAccount"`
}

// IBANAccount represents an IBAN account.
type IBANAccount struct {
	IBAN string `json:"iban"`
}

// CurrencyExchange represents a currency exchange value.
type CurrencyExchange struct {
	SourceCurrency string `json:"sourceCurrency"`
	ExchangeRate   string `json:"exchangeRate"`
	UnitCurrency   string `json:"unitCurrency"`
	TargetCurrency string `json:"targetCurrency"`
	QuotationDate  string `json:"quotationDate"`
}

// TransactionAmountValue is a value for a transaction amount.
type TransactionAmountValue float64

// UnmarshalJSON provides custom unmarshalling for TransactionAmountValue.
func (tav *TransactionAmountValue) UnmarshalJSON(b []byte) (err error) {
	amount, err := strconv.ParseFloat(strings.Trim(string(b), "\""), 64)
	if err != nil {
		return errors.Wrapf(err, "could not parse transaction amount value %q", strings.Trim(string(b), "\""))
	}

	*tav = TransactionAmountValue(amount)

	return nil
}

// Transactions returns a list of all transactions for the given account.
func (c *Client) Transactions(ctx context.Context, accountID string) (*Transactions, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/accounts/%s/transactions/", baseURL, accountID),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", "Bearer "+c.accessKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make http request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received unexpected status code when fetching transactions: %s", resp.Status)
	}

	var transactions struct {
		Transactions *Transactions `json:"transactions"`
	}

	err = json.NewDecoder(resp.Body).Decode(&transactions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response body")
	}

	return transactions.Transactions, nil
}
