package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/saucam/airflow-runner/cmd"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	cmd.Execute()
}
