package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/chia-network/go-chia-libs/pkg/types"
	log "github.com/sirupsen/logrus"
)

func (h *Healthcheck) farmerReceive(resp *types.WebsocketResponse) {
	if resp.Command != "new_signage_point" {
		return
	}

	newSP := &types.EventFarmerNewSignagePoint{}
	err := json.Unmarshal(resp.Data, newSP)
	if err != nil {
		log.Errorf("Error unmarshalling: %s\n", err.Error())
		return
	}

	h.lastFarmerTime = time.Now()
}

// farmerHealthcheck endpoint for the farmer service as a whole
func (h *Healthcheck) farmerHealthcheck() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timeMetricHealthcheckHelper(h.lastFarmerTime, w, r)
	}
}
