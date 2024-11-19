package gcpcollector

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/nianticlabs/modron/src/constants"
)

var meter = otel.Meter("github.com/nianticlabs/modron/src/collector/gcpcollector")

type metrics struct {
	SccCollectedObservations metric.Int64Counter
}

func initMetrics() metrics {
	sccCollectedObsCounter, err := meter.Int64Counter(constants.MetricsPrefix+"scc_collected_observations",
		metric.WithDescription("Number of collected observations from SCC"),
	)
	if err != nil {
		log.Errorf("failed to create scc_collected_observations counter: %v", err)
	}
	return metrics{
		SccCollectedObservations: sccCollectedObsCounter,
	}
}
