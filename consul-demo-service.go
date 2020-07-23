package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	promRequestsRoot = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "requests_root_total",
		Help: "Number of request to / endpoint.",
	})
	promRequestsLivez = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "requests_livez_total",
		Help: "Number of request to /livez endpoint.",
	})
	promRequestsRefresh = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "requests_refresh_total",
		Help: "Number of request to /refresh endpoint.",
	})
	promRefresh = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "refresh_total",
		Help: "Number of all config refreshs (time based + api).",
	})
)

type state struct {
	message string
}

func reloadConfig(s *state, kv *api.KV, region string, serviceID string) {
	pair, _, _ := kv.Get("config/global/message", nil)
	if pair != nil {
		s.message = string(pair.Value)
	}
	if region != "" {
		pair, _, _ = kv.Get("config/region/"+region+"/message", nil)
		if pair != nil {
			s.message = string(pair.Value)
		}
	}
	pair, _, _ = kv.Get("config/service/"+serviceID+"/message", nil)
	if pair != nil {
		s.message = string(pair.Value)
	}
	promRefresh.Inc()
}

func registerServiceWithConsul(client *api.Client, host string, port int, serviceID string) {
	client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      "consul-demo-service-" + serviceID,
		Name:    "consul-demo-service",
		Address: host,
		Port:    port,
		Tags:    []string{"prometheus"},
		Check: &api.AgentServiceCheck{
			HTTP:     "http://" + host + ":" + strconv.Itoa(port) + "/livez",
			Interval: "5s",
			Timeout:  "3s",
		},
	})
}

func getInstance() string {
	instance := os.Getenv("INSTANCE")
	if instance == "" {
		return "0"
	}
	return instance

}

func getRegion() string {
	region := os.Getenv("REGION")
	if region == "" {
		return "default"
	}
	return region
}

func getHost() string {
	host := os.Getenv("HOST")
	if host == "" {
		return "127.0.0.1"
	}
	return host
}

func getPort() int {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return 80
	}
	port, _ := strconv.Atoi(portStr)
	return port
}

func main() {
	instance := getInstance()
	region := getRegion()
	serviceID := region + "-" + instance
	host := getHost()
	port := getPort()
	s := state{"default"}

	prometheus.MustRegister(promRequestsRoot)
	prometheus.MustRegister(promRequestsLivez)
	prometheus.MustRegister(promRequestsRefresh)
	prometheus.MustRegister(promRefresh)

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}
	registerServiceWithConsul(client, host, port, serviceID)

	kv := client.KV()

	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("[main] Reload config (every 30s)")
				reloadConfig(&s, kv, region, serviceID)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "["+serviceID+"] ")
		fmt.Fprintf(w, s.message)
		fmt.Fprintf(w, "\n")
		promRequestsRoot.Inc()
		log.Println("[http] /")
	})

	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[main] Reload config (http)")
		reloadConfig(&s, kv, region, serviceID)
		fmt.Fprintf(w, "OK\n")
		promRequestsRefresh.Inc()
		log.Println("[http] /refresh")
	})

	http.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK\n")
		promRequestsLivez.Inc()
		log.Println("[http] /livez")
	})

	http.Handle("/metrics", promhttp.Handler())

	log.Println("[main] ServiceID:", serviceID)
	log.Println("[main] Port:", port)
	log.Println("[main] Reload config (startup)")
	reloadConfig(&s, kv, region, serviceID)

	server := &http.Server{Addr: ":" + strconv.Itoa(port)}
	go func() {
		log.Println("[main] Server started.")
		server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	client.Agent().ServiceDeregister("consul-demo-service-" + serviceID)
	log.Println("[main] Server terminated.")
}
