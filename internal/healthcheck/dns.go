package healthcheck

import (
	"context"
	"fmt"
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
				h.dnsOK = false
				return
			}

			if len(ips) > 0 {
				log.Println("Received at least 1 IP. Ready!")
				h.dnsOK = true
				return
			}

			log.Println("Received NO IPs. Not Ready!")
			h.dnsOK = false
		}()

		time.Sleep(30 * time.Second)
	}
}

// seederHealthcheck endpoint for the seeder service as a whole (Are we sending DNS responses)
func (h *Healthcheck) seederHealthcheck() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.dnsOK {
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
