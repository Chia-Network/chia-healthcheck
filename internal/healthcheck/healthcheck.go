package healthcheck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-chia-libs/pkg/types"
)

type Healthcheck struct {
	healthcheckPort uint16
	client          *rpc.Client

	// Last block height we received
	lastHeight uint32

	// Time we received the last block height
	lastHeightTime time.Time
}

func NewHealthcheck(port uint16, logLevel log.Level) (*Healthcheck, error) {
	var err error

	healthcheck := &Healthcheck{
		healthcheckPort: port,
	}

	log.SetLevel(logLevel)

	healthcheck.client, err = rpc.NewClient(rpc.ConnectionModeWebsocket, rpc.WithAutoConfig(), rpc.WithBaseURL(&url.URL{
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

	err = h.client.AddHandler(h.websocketReceive)
	if err != nil {
		return err
	}

	h.client.AddDisconnectHandler(h.disconnectHandler)
	h.client.AddReconnectHandler(h.reconnectHandler)

	return nil
}

// CloseWebsocket closes the websocket connection
func (h *Healthcheck) CloseWebsocket() error {
	//return m.client.DaemonService.CloseConnection()
	return nil
}

// StartServer starts the metrics server
func (h *Healthcheck) StartServer() error {
	log.Printf("Starting healthcheck server on port %d", h.healthcheckPort)

	http.HandleFunc("/full_node", h.fullNodeHealthcheck())
	return http.ListenAndServe(fmt.Sprintf(":%d", h.healthcheckPort), nil)
}

func (h *Healthcheck) websocketReceive(resp *types.WebsocketResponse, err error) {
	if err != nil {
		log.Errorf("Websocket received err: %s\n", err.Error())
		return
	}

	log.Printf("recv: %s %s\n", resp.Origin, resp.Command)
	log.Debugf("origin: %s command: %s destination: %s data: %s\n", resp.Origin, resp.Command, resp.Destination, string(resp.Data))

	if resp.Origin != "chia_full_node" {
		return
	}

	if resp.Command != "block" {
		return
	}

	block := &types.BlockEvent{}
	err = json.Unmarshal(resp.Data, block)
	if err != nil {
		log.Errorf("Error unmarshalling: %s\n", err.Error())
		return
	}

	h.lastHeight = block.Height
	h.lastHeightTime = time.Now()
}

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

// Healthcheck endpoint for the healthcheck service as a whole
func (h *Healthcheck) fullNodeHealthcheck() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if time.Now().Sub(h.lastHeightTime) < viper.GetDuration("healthcheck-threshold") {
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
}
