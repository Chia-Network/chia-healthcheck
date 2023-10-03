package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

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

// Healthcheck endpoint for the full node service as a whole
func (h *Healthcheck) fullNodeHealthcheck() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timeMetricHealthcheckHelper(h.lastHeightTime, w, r)
	}
}
