package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	metricsPath     = "/metrics"
	port            = "exporter.port"
	portDefault     = "8080"
	interval        = "exporter.collect.interval.ms"
	intervalDefault = 30000
)

var (
	missingServiceMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "swarm_service_replicas_missing",
			Help: "Shows whether cluster contains any service configuration without corresponding replica instances",
		},
		[]string{"service_name"},
	)
)

func main() {
	defineParam(port, "EXPORTER_PORT", portDefault)
	defineParam(interval, "EXPORTER_INTERVAL", intervalDefault)
	logrus.SetLevel(logrus.DebugLevel)

	prometheus.MustRegister(missingServiceMetric)
	gatherFunc()

	http.Handle(metricsPath, promhttp.Handler())
	//http.ListenAndServe(":"+viper.GetString(port), nil)
}

func gatherFunc() {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("Error creating Docker client")
		return
	}

	list, err := cli.TaskList(ctx, types.TaskListOptions{})
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("Error fetching service list")
		return
	}
	for _, task := range list {
		obj, _ := json.MarshalIndent(task, "", "  ")
		//logrus.Infof("service: %v ", string(obj))
		os.Stdout.Write(obj)
	}

	/* go func() {
		for {
			// refactor me
			//ctx := context.Background()
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				logrus.WithFields(logrus.Fields{"error": err}).Error("Error creating Docker client")
				return
			}
			list, err := cli.ServiceList(context.Background(), types.ServiceListOptions{})
			if err != nil {
				logrus.WithFields(logrus.Fields{"error": err}).Error("Error fetching service list")
				return
			}
			for _, srv := range list {
				srv.
			}

			// stub
			missingServiceMetric.WithLabelValues("some_service_name").Set(rand.Float64())
			time.Sleep(time.Duration(viper.GetInt(interval)) * time.Millisecond)
		}
	}() */
}

func defineParam(name, envName string, defaultValue interface{}) {
	viper.BindEnv(name, envName)
	viper.SetDefault(name, defaultValue)
}
