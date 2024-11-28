package healthcheck

import (
	"context"
	"net"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// DNSCheckLoop runs a loop checking for DNS responses
func (h *Healthcheck) DNSCheckLoop() {
	hostname := viper.GetString("dns-hostname")
	if len(hostname) == 0 {
		log.Println("dns-hostname not set. Skipping DNS Monitoring")
		return
	}

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 30 * time.Second,
			}
			return d.DialContext(ctx, network, "127.0.0.1:53")
		},
	}

	for {
		func() {
			ips, err := r.LookupIP(context.TODO(), "ip", hostname)
			if err != nil {
				log.Printf("Fetching dns records failed: %s\n", err.Error())
				return
			}

			if len(ips) > 0 {
				log.Println("Received at least 1 IP. Ready!")
				h.lastDNSTime = time.Now()
				return
			}

			log.Println("Received NO IPs. Not Ready!")
		}()

		time.Sleep(min(30*time.Second, viper.GetDuration("healthcheck-threshold")/2))
	}
}

// seederHealthcheck endpoint for the seeder service as a whole (Are we sending DNS responses)
func (h *Healthcheck) seederHealthcheck() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timeMetricHealthcheckHelper(h.lastDNSTimeGT1, w, r)
	}
}

func (h *Healthcheck) seederReadiness() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		timeMetricHealthcheckHelper(h.lastDNSTime, w, r)
	}
}
