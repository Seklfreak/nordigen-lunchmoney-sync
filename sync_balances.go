package main

import (
	"context"
	"strings"

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
	assets, err := lunchmoneyClient.GetAssets(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to fetch asses from Lunchmoney")
	}

	var asset *lunchmoney.Asset

	for _, a := range assets {
		if a.ID == lunchmoneyAssetID {
			asset = a
		}
	}

	if asset == nil {
		return errors.New("unable to find Lunchmoney asset to sync")
	}

	balances, err := nordigenClient.GetAccountBalances(ctx, nordigenAccountID)
	if err != nil {
		return errors.Wrap(err, "failed to fetch account balances from Nordigen")
	}

	var balance *nordigen.Balance

	for _, bl := range balances {
		if bl.BalanceType == "expected" && strings.EqualFold(bl.BalanceAmount.Currency, asset.Currency) {
			balance = bl

			break
		}
	}

	if balance == nil {
		return errors.New("unable to find a balance to sync on Nordigen")
	}

	lmBalance := lunchmoney.AssetBalance(balance.BalanceAmount.Amount)

	err = lunchmoneyClient.UpdateAsset(ctx, lunchmoneyAssetID, &lunchmoney.Asset{
		Balance: &lmBalance,
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
