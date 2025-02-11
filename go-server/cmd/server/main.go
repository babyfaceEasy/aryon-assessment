package main

import (
	"log/slog"

	"connector-recruitment/go-server/connectors/config"
	"connector-recruitment/go-server/connectors/db"
	"connector-recruitment/go-server/connectors/logger"
)

func main() {
	var env config.Env
	if err := config.LoadEnv(&env); err != nil {
		panic(err)
	}

	// setup logger
	slogger := logger.NewProductionLogger(env)
	slog.SetDefault(slogger)

	// get DB
	db := db.NewDB(env)
	// setup AWS session for LocalStack
	smClient := config.NewSecretClient(env)

	slogger.Info("successfully connected to postgres and has run migrations")
	grpcServer := NewGRPCServer(env.RPCPort, smClient, slogger, db)
	if err := grpcServer.Run(env); err != nil {
		slogger.Error("failed to serve: ", "err", err)
	}
}
