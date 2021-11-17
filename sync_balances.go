package main

import (
	"context"

	"github.com/Seklfreak/nordigen-lunchmoney-sync/lunchmoney"
	"github.com/Seklfreak/nordigen-lunchmoney-sync/nordigen"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func syncBalance(
	ctx context.Context,
	nordigenAccountID string,
	lunchmoneyAssetID int,
	nordigenClient *nordigen.Client,
	lunchmoneyClient *lunchmoney.Client,
	log *zap.Logger,
) error {
	// TODO: compare currencies to avoid syncing mismatching currencies

	balances, err := nordigenClient.GetAccountBalances(ctx, nordigenAccountID)
	if err != nil {
		return errors.Wrap(err, "failed to fetch account balances from Nordigen")
	}

	var balance *nordigen.Balance
	for _, bl := range balances {
		if bl.BalanceType == "expected" {
			balance = bl
			break
		}
	}

	if balance == nil {
		return errors.New("unable to find a balance to sync")
	}

	err = lunchmoneyClient.UpdateAsset(ctx, lunchmoneyAssetID, &lunchmoney.Asset{
		Balance: lunchmoney.AssetBalance(balance.BalanceAmount.Amount),
	})
	if err != nil {
		return errors.Wrap(err, "failed to update Lunchmoney asset")
	}

	log.Info("synced balance",
		zap.Float64("amount", float64(balance.BalanceAmount.Amount)),
		zap.String("nordigen_account_id", nordigenAccountID),
		zap.Int("lunchmoney_asset_id", lunchmoneyAssetID),
	)

	return nil
}
