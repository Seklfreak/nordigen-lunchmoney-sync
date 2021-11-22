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

	TransactionAmount Amount            `json:"transactionAmount"`
	CurrencyExchange  CurrencyExchanges `json:"currencyExchange"`

	BankTransactionCode                    string   `json:"bankTransactionCode"`
	AdditionalInformation                  string   `json:"additionalInformation"`
	RemittanceInformationUnstructured      string   `json:"remittanceInformationUnstructured"`
	RemittanceInformationUnstructuredArray []string `json:"remittanceInformationUnstructuredArray"`
	ProprietaryBankTransactionCode         string   `json:"proprietaryBankTransactionCode"`
	BookingDate                            Date     `json:"bookingDate"`
	ValueDate                              Date     `json:"valueDate"`
	UltimateCreditor                       string   `json:"ultimateCreditor"`
	MandateID                              string   `json:"mandateId"`
	EndToEndID                             string   `json:"endToEndId"`

	DebtorName    string       `json:"debtorName"`
	DebtorAccount *IBANAccount `json:"debtorAccount"`

	CreditorName    string       `json:"creditorName"`
	CreditorAccount *IBANAccount `json:"creditorAccount"`
}

// IBANAccount represents an IBAN account.
type IBANAccount struct {
	IBAN string `json:"iban"`
}

// CurrencyExchanges represents a collection of CurrencyExchange items.
type CurrencyExchanges []*CurrencyExchange

// UnmarshalJSON provides custom JSON unmarshalling for CurrencyExchanges.
func (ce *CurrencyExchanges) UnmarshalJSON(b []byte) error {
	// Nordigen returns this sometimes as an array of objects, sometimes as an object.
	// this logic is to handle both cases.

	// try array
	var r []*CurrencyExchange
	err := json.Unmarshal(b, &r)
	if err != nil {
		// try object
		var ceItem *CurrencyExchange
		err = json.Unmarshal(b, &ceItem)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal CurrencyExchanges")
		}

		r = []*CurrencyExchange{ceItem}
	}

	*ce = r

	return nil
}

// CurrencyExchange represents a currency exchange value.
type CurrencyExchange struct {
	SourceCurrency   string `json:"sourceCurrency"`
	ExchangeRate     string `json:"exchangeRate"`
	UnitCurrency     string `json:"unitCurrency"`
	TargetCurrency   string `json:"targetCurrency"`
	QuotationDate    string `json:"quotationDate"`
	InstructedAmount Amount `json:"instructedAmount"`
}

// TransactionAmountValue is a value for a transaction amount.
type TransactionAmountValue float64

// UnmarshalJSON provides custom JSON unmarshalling for TransactionAmountValue.
func (tav *TransactionAmountValue) UnmarshalJSON(b []byte) error {
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
