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
	lunchmoneyAssetID int,
	trxStatus lunchmoney.TransactionStatus,
) (*lunchmoney.Transaction, error) {
	payee := trx.CreditorName
	if payee == "" {
		payee = trx.DebtorName
	}

	lmTrx := &lunchmoney.Transaction{
		AssetID: lunchmoneyAssetID,

		Amount:     float64(trx.TransactionAmount.Amount),
		Currency:   strings.ToLower(trx.TransactionAmount.Currency),
		Date:       lunchmoney.TransactionDate(trx.ValueDate),
		Payee:      payee,
		Notes:      strings.Join(trx.RemittanceInformationUnstructuredArray, "; "),
		Status:     trxStatus,
		ExternalID: trx.TransactionID,
	}

	if lmTrx.AssetID <= 0 ||
		lmTrx.Amount == 0 ||
		lmTrx.Currency == "" ||
		time.Time(lmTrx.Date).IsZero() ||
		lmTrx.Payee == "" ||
		lmTrx.ExternalID == "" {
		return lmTrx, errors.New("created lunchmoney transaction failed validation")
	}

	return lmTrx, nil
}
