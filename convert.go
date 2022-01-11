package main

import (
	"crypto/sha256"
	"encoding/base64"
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

	// for exchange or transfer use the transaction code as payee
	if (trx.ProprietaryBankTransactionCode == "EXCHANGE" ||
		trx.ProprietaryBankTransactionCode == "TRANSFER") &&
		payee == "" {
		payee = strings.Title(strings.ToLower(trx.ProprietaryBankTransactionCode))
	}

	note := trx.RemittanceInformationUnstructured
	if note == "" {
		note = strings.Join(trx.RemittanceInformationUnstructuredArray, "; ")
	}

	transactionID := trx.TransactionID
	if transactionID == "" {
		// if API returns no external Transaction ID build new one out of hash of all information
		transactionID = fmt.Sprintf(
			"%s|%.2f%s|%s|%s|%s",
			time.Time(trx.ValueDate),
			trx.TransactionAmount.Amount,
			trx.TransactionAmount.Currency,
			trx.CreditorName,
			trx.DebtorName,
			note,
		)

		hasher := sha256.New()
		hasher.Write([]byte(transactionID))
		transactionID = base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	}

	lmTrx := &lunchmoney.Transaction{
		AssetID: lunchmoneyAssetID,

		Amount:     float64(trx.TransactionAmount.Amount),
		Currency:   strings.ToLower(trx.TransactionAmount.Currency),
		Date:       lunchmoney.TransactionDate(date),
		Payee:      payee,
		Notes:      note,
		Status:     lunchmoney.TransactionStatusUncleared,
		ExternalID: transactionID,

		Tags: []string{"nordigen-lunchmoney-sync"},
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
