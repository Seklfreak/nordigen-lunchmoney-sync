package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Seklfreak/nordigen-lunchmoney-sync/lunchmoney"
	"github.com/Seklfreak/nordigen-lunchmoney-sync/nordigen"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func main() {
	// init logger
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(errors.Wrap(err, "failed to create logger"))
	}
	defer log.Sync()
	zap.ReplaceGlobals(log)

	// parse config
	var config struct {
		Nordigen          *nordigen.Config `envconfig:"NORDIGEN" required:"true"`
		NordigenAccountID string           `envconfig:"NORDIGEN_ACCOUNT_ID" required:"true"`

		LunchmoneyAccessToken string `envconfig:"LUNCHMONEY_ACCESS_TOKEN" required:"true"`
		LunchmoneyAssetID     int    `envconfig:"LUNCHMONEY_ASSET_ID" required:"true"`
	}
	err = envconfig.Process("", &config)
	if err != nil {
		log.Fatal("failed to process config", zap.Error(err))
	}

	// create Nordigen client
	nordigenClient, err := nordigen.NewClient(
		config.Nordigen,
		&http.Client{
			Timeout: 60 * time.Second,
		},
	)
	if err != nil {
		log.Fatal("failed to create nordigen client", zap.Error(err))
	}

	// create Lunchmoney client
	lunchmoneyClient := lunchmoney.NewClient(
		config.LunchmoneyAccessToken,
		&http.Client{
			Timeout: 60 * time.Second,
		},
	)

	ctx := context.Background()

	// fetch transactions from Nordigen
	transactions, err := nordigenClient.Transactions(ctx, config.NordigenAccountID)
	if err != nil {
		log.Fatal("failed to fetch transactions from nordigen", zap.Error(err))
	}

	allTransactions := append(transactions.Booked, transactions.Pending...)
	log.Info("fetched transactions from Nordigen", zap.Int("total", len(allTransactions)))

	// prepare transactions to insert
	lunchmoneyTransactions := make([]*lunchmoney.Transaction, 0, len(allTransactions))

	for _, trx := range allTransactions {
		payee := trx.CreditorName
		if payee == "" {
			payee = trx.DebtorName
		}

		lmTrx := &lunchmoney.Transaction{
			AssetID: config.LunchmoneyAssetID,

			Amount:     float64(trx.TransactionAmount.Amount),
			Currency:   strings.ToLower(trx.TransactionAmount.Currency),
			Date:       lunchmoney.TransactionDate(trx.ValueDate),
			Payee:      payee,
			Notes:      strings.Join(trx.RemittanceInformationUnstructuredArray, "; "),
			Status:     lunchmoney.TransactionStatusCleared, // TODO: set uncleared for pending transactions
			ExternalID: trx.TransactionID,
		}

		if lmTrx.AssetID <= 0 ||
			lmTrx.Amount == 0 ||
			lmTrx.Currency == "" ||
			time.Time(lmTrx.Date).IsZero() ||
			lmTrx.Payee == "" ||
			lmTrx.ExternalID == "" {
			log.Fatal("unable to create proper lunchmoney transaction",
				zap.Any("lunchmoney_trx", lmTrx),
				zap.Any("source_transaction", trx),
			)
		}

		lunchmoneyTransactions = append(lunchmoneyTransactions, lmTrx)
	}

	for _, trx := range lunchmoneyTransactions {
		log.Info("prepared transaction", zap.Any("transaction", trx))
	}

	// split all into chunks
	chunkSize := 50
	chunks := make([][]*lunchmoney.Transaction, 0, chunkSize)

	for i := 0; i < len(lunchmoneyTransactions); i += chunkSize {
		end := i + chunkSize

		if end > len(lunchmoneyTransactions) {
			end = len(lunchmoneyTransactions)
		}

		chunks = append(chunks, lunchmoneyTransactions[i:end])
	}

	// insert transactions
	for _, chunk := range chunks {
		inserted, err := lunchmoneyClient.InsertTransactions(ctx, chunk)
		if err != nil {
			log.Fatal("failed to insert transactions", zap.Error(err), zap.Int("inserted_count", inserted), zap.Int("chunk_size", len(chunk)))
		}

		log.Info("inserted transactions", zap.Int("inserted_count", inserted), zap.Int("chunk_size", len(chunk)))
	}
}
