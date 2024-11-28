package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/chia-network/go-chia-libs/pkg/types"
)

func (h *Healthcheck) fullNodeReceive(resp *types.WebsocketResponse) {
	var blockHeight uint32

	if resp.Command != "get_blockchain_state" {
		return
	}

	block := &types.WebsocketBlockchainState{}
	err := json.Unmarshal(resp.Data, block)
	if err != nil {
		log.Errorf("Error unmarshalling: %s\n", err.Error())
		return
	}
	blockHeight = block.BlockchainState.Peak.OrEmpty().Height

	// Edge case, but we should be sure block height is increasing
	if blockHeight <= h.lastHeight {
		return
	}

	h.lastHeight = blockHeight
	h.lastHeightTime = time.Now()
}

// FullNodeCheckLoop runs a loop checking if full node ports are open
func (h *Healthcheck) FullNodeCheckLoop() {
	for {
		func() {
			if !isPortOpen(viper.GetString("hostname"), h.chiaConfig.FullNode.Port) {
				log.Errorf("Full node port %d is not open", h.chiaConfig.FullNode.Port)
				return
			}
			if !isPortOpen(viper.GetString("hostname"), h.chiaConfig.FullNode.RPCPort) {
				log.Errorf("Full node RPC port %d is not open", h.chiaConfig.FullNode.RPCPort)
				return
			}
			h.lastFullNodeActivity = time.Now()
		}()

		// Loop every thirty seconds, or healthcheckthreshold/2 if the threshold is less than 15seconds
		time.Sleep(min(30*time.Second, viper.GetDuration("healthcheck-threshold")/2))
	}
}

// Healthcheck endpoint for the full node service as a whole
func (h *Healthcheck) fullNodeHealthcheck() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timeMetricHealthcheckHelper(h.lastHeightTime, w, r)
	}
}

// Healthcheck endpoint for the full node service as a whole
func (h *Healthcheck) fullNodeReadiness() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timeMetricHealthcheckHelper(h.lastFullNodeActivity, w, r)
	}
}
