package nordigen

import (
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Amount represents a currency amount.
type Amount struct {
	Amount   TransactionAmountValue `json:"amount"`
	Currency string                 `json:"currency"`
}

// Date represents a date without a time.
type Date time.Time

// UnmarshalJSON provides custom unmarshalling for Date.
func (td *Date) UnmarshalJSON(b []byte) (err error) {
	transactionDate, err := time.Parse("2006-01-02", strings.Trim(string(b), "\""))
	if err != nil {
		return errors.Wrap(err, "could not parse transaction date")
	}

	*td = Date(transactionDate)

	return nil
}
