package cmd

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/chia-healthcheck/internal/healthcheck"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the healthcheck server",
	Run: func(cmd *cobra.Command, args []string) {
		level, err := log.ParseLevel(viper.GetString("log-level"))
		if err != nil {
			log.Fatalf("Error parsing log level: %s\n", err.Error())
		}

		h, err := healthcheck.NewHealthcheck(uint16(viper.GetInt("healthcheck-port")), level)
		if err != nil {
			log.Fatalln(err.Error())
		}

		// Run this in the background, so the metrics healthz endpoint can come up while waiting for Chia
		go startWebsocket(h)
		go h.DNSCheckLoop()

		log.Fatalln(h.StartServer())
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func startWebsocket(h *healthcheck.Healthcheck) {
	// Loop until we get a connection or cancel
	// This enables starting the healthcheck app even if the chia RPC service is not up/responding
	// It just retries every 5 seconds to connect to the RPC server until it succeeds or the app is stopped
	for {
		err := h.OpenWebsocket()
		if err != nil {
			log.Println(err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
}
