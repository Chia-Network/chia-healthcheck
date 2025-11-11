package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/chia-network/go-chia-libs/pkg/types"
	log "github.com/sirupsen/logrus"
)

func (h *Healthcheck) harvesterReceive(resp *types.WebsocketResponse) {
	if resp.Command != "farming_info" {
		return
	}

	farmingInfo := &types.EventHarvesterFarmingInfo{}
	err := json.Unmarshal(resp.Data, farmingInfo)
	if err != nil {
		log.Errorf("Error unmarshalling: %s\n", err.Error())
		return
	}

	h.lastHarvesterTime = time.Now()

	if farmingInfo.TotalPlots == 0 {
		log.Errorf("No plots found. Not Ready!")
		return
	}

	h.lastHarvesterTimeWithPlots = time.Now()
}

// harvesterHealthcheck endpoint for the harvester service as a whole
func (h *Healthcheck) harvesterHealthcheckWithPlots() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timeMetricHealthcheckHelper(h.lastHarvesterTimeWithPlots, w, r)
	}
}

// timelordHealthcheck endpoint for the timelord service as a whole
func (h *Healthcheck) harvesterHealthcheck() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timeMetricHealthcheckHelper(h.lastHarvesterTime, w, r)
	}
}
