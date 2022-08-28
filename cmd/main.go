package main

import (
	"context"
	"os"
	"os/signal"

	"gitlab.com/g6834/team31/auth/pkg/logging"
	"gitlab.com/g6834/team31/mailsender/internal/adapters/mq"
	"gitlab.com/g6834/team31/mailsender/internal/config"
	"gitlab.com/g6834/team31/mailsender/internal/domain/sender"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.NewConfig()
	logger := logging.New(cfg.Log.Level)
	logger.Info().Msgf("config: %+v", cfg)
	service := sender.New(cfg.Sender, &logger)
	mq := mq.BuildMQClietn(cfg.Kafka, service, &logger)
	producerErrChan := service.LimitedSend(ctx)
	logger.Info().Msg("launching mq client")
	mqErrChan := mq.RunMQ(ctx, producerErrChan)
	mq.RunMQ(ctx, producerErrChan)
	service.LimitedSend(ctx)
	osInterupt := make(chan os.Signal, 1)
	signal.Notify(osInterupt, os.Interrupt)
	select {
	case <-osInterupt:
		cancel()
	case err := <-mqErrChan:
		logger.Fatal().Err(err).Msg("kafka bad")
	}
}
