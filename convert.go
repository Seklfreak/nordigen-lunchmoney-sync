package main

import (
	"errors"
	"strings"
	"time"

	"github.com/Seklfreak/nordigen-lunchmoney-sync/lunchmoney"
	"github.com/Seklfreak/nordigen-lunchmoney-sync/nordigen"
)

func createLunchmoneyTrx(
	trx nordigen.Transaction,
	account *nordigen.Account,
	lunchmoneyAssetID int,
	trxStatus lunchmoney.TransactionStatus,
) (*lunchmoney.Transaction, error) {
	payee := trx.CreditorName
	if payee == "" {
		payee = trx.DebtorName
	}

	date := trx.ValueDate
	if time.Time(date).IsZero() {
		date = trx.BookingDate
	}

	// for transfers from wallets to personal bank account use OwnerName as payee (e.g. PayPal)
	if trx.AdditionalInformation == "MONEY_TRANSFER" && payee == "" && account.OwnerName != "" {
		payee = account.OwnerName
	}

	lmTrx := &lunchmoney.Transaction{
		AssetID: lunchmoneyAssetID,

		Amount:     float64(trx.TransactionAmount.Amount),
		Currency:   strings.ToLower(trx.TransactionAmount.Currency),
		Date:       lunchmoney.TransactionDate(date),
		Payee:      payee,
		Notes:      strings.Join(trx.RemittanceInformationUnstructuredArray, "; "),
		Status:     trxStatus,
		ExternalID: trx.TransactionID,
	}

	if lmTrx.AssetID <= 0 {
		return nil, errors.New("lunchmoney transaction asset id cannot be empty")
	}

	if lmTrx.Amount == 0 {
		return nil, errors.New("lunchmoney transaction amount cannot be 0")
	}

	if lmTrx.Currency == "" {
		return nil, errors.New("lunchmoney transaction currency cannot be empty")
	}

	if time.Time(lmTrx.Date).IsZero() {
		return nil, errors.New("lunchmoney transaction date cannot be empty")
	}

	if lmTrx.Payee == "" {
		return nil, errors.New("lunchmoney transaction payee cannot be empty")
	}

	if lmTrx.ExternalID == "" {
		return nil, errors.New("lunchmoney transaction external ID cannot be empty")
	}

	return lmTrx, nil
}
