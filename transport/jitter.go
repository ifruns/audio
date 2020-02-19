package transport

import "time"

// Jitter buffer for RTP.
type Jitter struct {
	timer time.Timer

	latency float32
}
