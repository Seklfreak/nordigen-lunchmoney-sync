package main

import (
	"context"
	"fmt"

	"github.com/Seklfreak/nordigen-lunchmoney-sync/lunchmoney"
	"github.com/Seklfreak/nordigen-lunchmoney-sync/nordigen"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func printAccounts(
	ctx context.Context,
	nordigenRequisitionIDs []string,
	nordigenClient *nordigen.Client,
	lunchmoneyClient *lunchmoney.Client,
	log *zap.Logger,
) error {
	for _, nordigenRequisitionID := range nordigenRequisitionIDs {
		nordigenAccountList, err := nordigenClient.ListAccounts(ctx, nordigenRequisitionID)
		if err != nil {
			return errors.Wrapf(err, "failed to list accounts for requestion ID %q", nordigenRequisitionID)
		}

		for _, nordigenAccountID := range nordigenAccountList.Accounts {
			nordigenAccountDetails, err := nordigenClient.GetAccountDetails(ctx, nordigenAccountID)
			if err != nil {
				log.Warn("failed to fetch account details for account ID",
					zap.Error(err),
					zap.String("account_id", nordigenAccountID),
				)

				continue
			}

			nordigenAccountBalances, err := nordigenClient.GetAccountBalances(ctx, nordigenAccountID)
			if err != nil {
				log.Warn("failed to fetch account balances for account ID",
					zap.Error(err),
					zap.String("account_id", nordigenAccountID),
				)

				continue
			}

			balances := make(map[string]string)

			for _, balance := range nordigenAccountBalances {
				balances[balance.BalanceType] = fmt.Sprintf(
					"%.2f %s",
					balance.BalanceAmount.Amount,
					balance.BalanceAmount.Currency,
				)
			}

			log.Info("nordigen account",
				zap.String("id", nordigenAccountID),
				zap.String("name", nordigenAccountDetails.Name),
				zap.String("product", nordigenAccountDetails.Product),
				zap.String("status", nordigenAccountDetails.Status),
				zap.Any("balances", balances),
			)
		}
	}

	accounts, err := lunchmoneyClient.GetAssets(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to fetch lunchmoney accounts")
	}

	for _, account := range accounts {
		log.Info("lunchmoney account",
			zap.Int("id", account.ID),
			zap.String("name", account.Name),
			zap.String("institution_name", account.InstitutionName),
			zap.String("type", account.TypeName),
			zap.String("subtype", account.SubtypeName),
			zap.Float64("balance", float64(*account.Balance)),
			zap.String("currency", account.Currency),
		)
	}

	return nil
}
