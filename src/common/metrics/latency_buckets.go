package metrics

// LatencySecondsBuckets are second-scaled histogram bounds (the SDK defaults are ms-scaled, useless for seconds).
var LatencySecondsBuckets = []float64{
	0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10,
}

// LongLatencySecondsBuckets extend the tail for settlements that wait on a receipt.
var LongLatencySecondsBuckets = []float64{
	0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 20, 30, 60, 120,
}

// DbLatencySecondsBuckets are the DB semconv advisory set; a sub-millisecond redis GET lands in the first bucket.
var DbLatencySecondsBuckets = []float64{
	0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10,
}

// RpcLatencyMillisecondsBuckets are millisecond-scaled: rpc.client.duration is defined in ms.
var RpcLatencyMillisecondsBuckets = []float64{
	1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000, 30000, 120000,
}
