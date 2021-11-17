package main

import (
	"fmt"
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

	// for transfers from/to wallets using the personal account use OwnerName as payee (e.g. PayPal)
	if (trx.AdditionalInformation == "MONEY_TRANSFER" ||
		trx.ProprietaryBankTransactionCode == "TOPUP") &&
		payee == "" && account.OwnerName != "" {
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
		return nil, fmt.Errorf("converting trx %s: lunchmoney transaction asset id cannot be empty", trx.TransactionID)
	}

	if lmTrx.Amount == 0 {
		return nil, fmt.Errorf("converting trx %s: lunchmoney transaction amount cannot be 0", trx.TransactionID)
	}

	if lmTrx.Currency == "" {
		return nil, fmt.Errorf("converting trx %s: lunchmoney transaction currency cannot be empty", trx.TransactionID)
	}

	if time.Time(lmTrx.Date).IsZero() {
		return nil, fmt.Errorf("converting trx %s: lunchmoney transaction date cannot be empty", trx.TransactionID)
	}

	if lmTrx.Payee == "" {
		return nil, fmt.Errorf("converting trx %s: lunchmoney transaction payee cannot be empty", trx.TransactionID)
	}

	if lmTrx.ExternalID == "" {
		return nil, fmt.Errorf("converting trx %s: lunchmoney transaction external ID cannot be empty", trx.TransactionID)
	}

	return lmTrx, nil
}
