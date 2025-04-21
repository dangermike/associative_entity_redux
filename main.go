package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dangermike/associative_entity_redux/bench"
	"github.com/dangermike/associative_entity_redux/load"
	"github.com/dangermike/associative_entity_redux/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func main() {
	cmd := new(cobra.Command)
	cmd.AddCommand(load.Command(), bench.Command())

	log, err := zap.NewDevelopmentConfig().Build(zap.WithCaller(false))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create logger:", err)
		os.Exit(1)
	}

	ctx := logging.NewContext(context.Background(), log)

	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
