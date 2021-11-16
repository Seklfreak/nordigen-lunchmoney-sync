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
		Nordigen *nordigen.Config `envconfig:"NORDIGEN" required:"true"`

		LunchmoneyAccessToken string `envconfig:"LUNCHMONEY_ACCESS_TOKEN" required:"true"`

		Mapping map[string]int `envconfig:"MAPPING"` // map[nordigenAccountID]lunchmoneyAssetID
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
	if len(config.Mapping) == 0 {
		log.Info("no mapping found, printing accounts")

		accounts, err := lunchmoneyClient.GetAssets(ctx)
		if err != nil {
			log.Fatal("failed to fetch accounts from lunchmoney", zap.Error(err))
		}

		for _, account := range accounts {
			log.Info("lunchmoney account",
				zap.Int("id", account.ID),
				zap.String("name", account.Name),
				zap.String("institution_name", account.InstitutionName),
				zap.String("type", account.TypeName),
				zap.String("subtype", account.SubtypeName),
				zap.String("balance", account.Balance),
				zap.String("currency", account.Currency),
			)
		}

		return
	}

	for nordigenAccountID, lunchmoneyAssetID := range config.Mapping {
		err = syncAccount(
			ctx,
			nordigenAccountID,
			lunchmoneyAssetID,
			nordigenClient,
			lunchmoneyClient,
			log,
		)
		if err != nil {
			log.Fatal("failure syncing account",
				zap.String("nordigen_account_id", nordigenAccountID),
				zap.Int("lunchmoney_asset_id", lunchmoneyAssetID),
				zap.Error(err),
			)
		}
	}
}
