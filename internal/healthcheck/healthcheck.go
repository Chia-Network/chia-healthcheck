package healthcheck

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/chia-network/go-chia-libs/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-chia-libs/pkg/types"
)

// Healthcheck is the main container for the app
type Healthcheck struct {
	healthcheckPort uint16
	client          *rpc.Client
	chiaConfig      *config.ChiaConfig

	// Last time the full node was synced
	lastSyncedTime time.Time

	// Last block height we received
	lastHeight uint32

	// Time we received the last block height
	lastHeightTime time.Time

	// Last full node activity
	lastFullNodeActivity time.Time

	// Last time we got a successful DNS response
	lastDNSTime time.Time
	// Last time we got a successful DNS response with at least one peer
	lastDNSTimeGT1 time.Time

	// Time we got a good response from the timelord
	lastTimelordTime time.Time
}

// NewHealthcheck returns a new instance of healthcheck
func NewHealthcheck(port uint16, logLevel log.Level) (*Healthcheck, error) {
	var err error

	healthcheck := &Healthcheck{
		healthcheckPort: port,
	}

	log.SetLevel(logLevel)

	chiaConfig, err := config.GetChiaConfig()
	if err != nil {
		return nil, err
	}
	healthcheck.chiaConfig = chiaConfig
	healthcheck.client, err = rpc.NewClient(rpc.ConnectionModeWebsocket, rpc.WithManualConfig(*chiaConfig), rpc.WithBaseURL(&url.URL{
		Scheme: "wss",
		Host:   viper.GetString("hostname"),
	}))
	if err != nil {
		return nil, err
	}

	return healthcheck, nil
}

// OpenWebsocket sets up the RPC client and subscribes to relevant topics
func (h *Healthcheck) OpenWebsocket() error {
	err := h.client.Subscribe("metrics")
	if err != nil {
		return err
	}

	_, err = h.client.AddHandler(h.websocketReceive)
	if err != nil {
		return err
	}

	h.client.AddDisconnectHandler(h.disconnectHandler)
	h.client.AddReconnectHandler(h.reconnectHandler)

	return nil
}

// StartServer starts the metrics server
func (h *Healthcheck) StartServer() error {
	log.Printf("Starting healthcheck server on port %d", h.healthcheckPort)

	http.HandleFunc("/full_node", h.fullNodeHealthcheck())
	http.HandleFunc("/full_node/startup", h.fullNodeStartup())
	http.HandleFunc("/full_node/readiness", h.fullNodeReadiness())
	http.HandleFunc("/full_node/liveness", h.fullNodeLiveness())
	http.HandleFunc("/full_node/ports", h.fullNodePorts())
	http.HandleFunc("/seeder", h.seederHealthcheck())
	http.HandleFunc("/seeder/readiness", h.seederReadiness())
	http.HandleFunc("/timelord", h.timelordHealthcheck())
	http.HandleFunc("/timelord/readiness", h.timelordHealthcheck())
	return http.ListenAndServe(fmt.Sprintf(":%d", h.healthcheckPort), nil)
}

func (h *Healthcheck) websocketReceive(resp *types.WebsocketResponse, err error) {
	if err != nil {
		log.Errorf("Websocket received err: %s\n", err.Error())
		return
	}

	log.Printf("recv: %s %s\n", resp.Origin, resp.Command)
	log.Debugf("origin: %s command: %s destination: %s data: %s\n", resp.Origin, resp.Command, resp.Destination, string(resp.Data))

	switch resp.Origin {
	case "chia_full_node":
		h.fullNodeReceive(resp)
	case "chia_wallet":
		h.walletReceive(resp)
	case "chia_crawler":
		h.crawlerReceive(resp)
	case "chia_timelord":
		h.timelordReceive(resp)
	case "chia_harvester":
		h.harvesterReceive(resp)
	case "chia_farmer":
		h.farmerReceive(resp)
	}
}

func (h *Healthcheck) walletReceive(resp *types.WebsocketResponse) {}

func (h *Healthcheck) crawlerReceive(resp *types.WebsocketResponse) {}

func (h *Healthcheck) harvesterReceive(resp *types.WebsocketResponse) {}

func (h *Healthcheck) farmerReceive(resp *types.WebsocketResponse) {}

func (h *Healthcheck) disconnectHandler() {
	log.Debug("Calling disconnect handlers")
	// @TODO should we mark unhealthy immediately?
}

func (h *Healthcheck) reconnectHandler() {
	log.Debug("Calling reconnect handlers")
	err := h.client.Subscribe("metrics")
	if err != nil {
		log.Errorf("Error subscribing to metrics events: %s\n", err.Error())
	}
}

func timeMetricHealthcheckHelper(lastTime time.Time, w http.ResponseWriter, r *http.Request) {
	if time.Since(lastTime) < viper.GetDuration("healthcheck-threshold") {
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintf(w, "Ok")
		if err != nil {
			log.Errorf("Error writing healthcheck response %s\n", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := fmt.Fprintf(w, "Not OK")
		if err != nil {
			log.Errorf("Error writing healthcheck response %s\n", err.Error())
		}
	}
}

func isPortOpen(host string, port uint16) bool {
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		// Port is not open or the host is unreachable
		return false
	}
	_ = conn.Close()
	return true
}
