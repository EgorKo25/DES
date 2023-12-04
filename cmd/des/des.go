package main

import (
	"context"

	"github.com/EgorKo25/DES/internal/logger"
	"github.com/EgorKo25/DES/internal/workers"
)

func main() {

	ctx := context.Background()

	channel := make(chan chan []byte)

	log := logger.NewLoggers()

	worker := workers.NewWorkerPull(ctx, channel,
		6, 8, 3,
		log[1].Sugar(), "", "")

	log[0].Info("SERVICE INIT SUCCESS")

}
