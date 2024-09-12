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

		var h *healthcheck.Healthcheck

		// Loop until we get a connection or cancel
		// It just retries every 5 seconds to connect to the RPC server until it succeeds or the app is stopped
		for {
			h, err = healthcheck.NewHealthcheck(uint16(viper.GetInt("healthcheck-port")), level)
			if err != nil {
				log.Fatalln(err.Error())
			}

			err = startWebsocket(h)
			if err != nil {
				log.Printf("error starting websocket. Creating new client and trying again in 5 seconds: %s\n", err.Error())
				time.Sleep(5 * time.Second)
				continue
			}

			go h.DNSCheckLoop()
			break
		}

		log.Fatalln(h.StartServer())
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func startWebsocket(h *healthcheck.Healthcheck) error {
	err := h.OpenWebsocket()
	if err != nil {
		return err
	}
	return nil
}
