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
		Nordigen          *nordigen.Config `envconfig:"NORDIGEN" required:"true"`
		NordigenAccountID string           `envconfig:"NORDIGEN_ACCOUNT_ID" required:"true"`

		LunchmoneyAccessToken string `envconfig:"LUNCHMONEY_ACCESS_TOKEN"`
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

	// print transactions
	for _, trx := range transactions.Booked {
		log.Info("booked transaction", zap.Any("transaction", trx), zap.Time("value_date", time.Time(trx.ValueDate)))
	}

	for _, trx := range transactions.Pending {
		log.Info("pending transaction", zap.Any("transaction", trx), zap.Time("value_date", time.Time(trx.ValueDate)))
	}

	// fetch assets from Lunchmoney
	assets, err := lunchmoneyClient.GetAssets(ctx)
	if err != nil {
		log.Fatal("failed to fetch assets from lunchmoney", zap.Error(err))
	}

	for _, asset := range assets {
		log.Info("asset", zap.Any("asset", asset))
	}
}
