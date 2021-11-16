package main

import (
	"context"
	"net/http"
	"time"

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

	// fetch transactions from Nordigen
	transactions, err := nordigenClient.Transactions(context.Background(), config.NordigenAccountID)
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
}
