package main

import (
	"context"

	"github.com/Seklfreak/nordigen-lunchmoney-sync/lunchmoney"
	"github.com/Seklfreak/nordigen-lunchmoney-sync/nordigen"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func syncAccount(
	ctx context.Context,
	nordigenAccountID string,
	lunchmoneyAssetID int,
	nordigenClient *nordigen.Client,
	lunchmoneyClient *lunchmoney.Client,
	log *zap.Logger,
) error {
	// fetch account details from Nordigen
	account, err := nordigenClient.GetAccountDetails(ctx, nordigenAccountID)
	if err != nil {
		return errors.Wrap(err, "failed to fetch account details from Nordigen")
	}

	// fetch transactions from Nordigen
	transactions, err := nordigenClient.Transactions(ctx, nordigenAccountID)
	if err != nil {
		return errors.Wrap(err, "failed to fetch transactions from Nordigen")
	}

	log.Info("fetched transactions from Nordigen", zap.Int("total", len(transactions.Booked)+len(transactions.Pending)))

	// prepare transactions to insert
	lunchmoneyTransactions := make([]*lunchmoney.Transaction, 0, len(transactions.Booked)+len(transactions.Pending))

	for _, trx := range transactions.Booked {
		lmTrx, err := createLunchmoneyTrx(trx, account, lunchmoneyAssetID, lunchmoney.TransactionStatusCleared)
		if err != nil {
			return errors.Wrapf(err, "failed to create Lunchmoney transaction for Nordigen transaction %s", trx.TransactionID)
		}

		lunchmoneyTransactions = append(lunchmoneyTransactions, lmTrx)
	}

	for _, trx := range transactions.Pending {
		lmTrx, err := createLunchmoneyTrx(trx, account, lunchmoneyAssetID, lunchmoney.TransactionStatusUncleared)
		if err != nil {
			return errors.Wrapf(err, "failed to create Lunchmoney transaction for Nordigen transaction %s", trx.TransactionID)
		}

		lunchmoneyTransactions = append(lunchmoneyTransactions, lmTrx)
	}

	for _, trx := range lunchmoneyTransactions {
		log.Debug("prepared transaction", zap.Any("transaction", trx))
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
			return errors.Wrapf(err, "failed to insert transactions")
		}

		log.Info("inserted transactions",
			zap.Int("inserted_count", inserted),
			zap.Int("chunk_size", len(chunk)),
			zap.String("nordigen_account_id", nordigenAccountID),
			zap.Int("lunchmoney_asset_id", lunchmoneyAssetID),
		)
	}

	return nil
}
