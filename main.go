package main

import (
	"context"
	"net/http"
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
		Nordigen               *nordigen.Config `envconfig:"NORDIGEN" required:"true"`
		NordigenRequisitionIDs []string         `envconfig:"NORDIGEN_REQUISITION_IDS"`

		LunchmoneyAccessToken string `envconfig:"LUNCHMONEY_ACCESS_TOKEN" required:"true"`

		TransactionsMap map[string]int `envconfig:"TRANSACTIONS_MAP"` // map[nordigenAccountID]lunchmoneyAssetID
		BalancesMap     map[string]int `envconfig:"BALANCES_MAP"`     // map[nordigenAccountID]lunchmoneyAssetID
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

	// print accounts if there is no mapping
	if len(config.TransactionsMap) == 0 && len(config.BalancesMap) == 0 {
		log.Info("no mapping found, printing accounts")

		err = printAccounts(ctx, config.NordigenRequisitionIDs, nordigenClient, lunchmoneyClient, log)
		if err != nil {
			log.Fatal("failed to print accounts", zap.Error(err))
		}

		return
	}

	for nordigenAccountID, lunchmoneyAssetID := range config.TransactionsMap {
		err = syncAccount(
			ctx,
			nordigenAccountID,
			lunchmoneyAssetID,
			nordigenClient,
			lunchmoneyClient,
			log,
		)
		if err != nil {
			log.Fatal("failure syncing transactions",
				zap.String("nordigen_account_id", nordigenAccountID),
				zap.Int("lunchmoney_asset_id", lunchmoneyAssetID),
				zap.Error(err),
			)
		}
	}

	for nordigenAccountID, lunchmoneyAssetID := range config.BalancesMap {
		err = syncBalance(
			ctx,
			nordigenAccountID,
			lunchmoneyAssetID,
			nordigenClient,
			lunchmoneyClient,
			log,
		)
		if err != nil {
			log.Fatal("failure syncing balance",
				zap.String("nordigen_account_id", nordigenAccountID),
				zap.Int("lunchmoney_asset_id", lunchmoneyAssetID),
				zap.Error(err),
			)
		}
	}
}
