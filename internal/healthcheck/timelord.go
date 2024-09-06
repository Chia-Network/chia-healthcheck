package healthcheck

import (
	"net/http"
	"time"

	"github.com/chia-network/go-chia-libs/pkg/types"
)

// timelordReceive gets timelord events
func (h *Healthcheck) timelordReceive(resp *types.WebsocketResponse) {
	switch resp.Command {
	case "finished_pot":
		h.lastTimelordTime = time.Now()
	case "skipping_peak":
		// Fastest timelord
		h.lastTimelordTime = time.Now()
	case "new_peak":
		// Slowest Timelord
	}
}

// seederHealthcheck endpoint for the seeder service as a whole (Are we sending DNS responses)
func (h *Healthcheck) timelordHealthcheck() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timeMetricHealthcheckHelper(h.lastTimelordTime, w, r)
	}
}
