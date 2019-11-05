package main

import (
	"context"
	"net/http"
	"time"

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
	taskStateMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "swarm_task_state_details",
			Help: "Shows state details for task instances running (per service configured)",
		},
		[]string{"service_name", "current_state", "desired_state"},
	)
)

func main() {
	defineParam(port, "METRICS_PORT", portDefault)
	defineParam(interval, "METRICS_INTERVAL", intervalDefault)
	logrus.SetLevel(logrus.DebugLevel)

	prometheus.MustRegister(taskStateMetric)
	gatherFunc()

	http.Handle(metricsPath, promhttp.Handler())
	http.ListenAndServe(":"+viper.GetString(port), nil)
}

func gatherFunc() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("Error creating Docker client")
		return
	}
	go func() {
		for {
			list, err := cli.TaskList(ctx, types.TaskListOptions{})
			if err != nil {
				logrus.WithFields(logrus.Fields{"error": err}).Error("Error fetching service list")
				return
			}
			for _, task := range list {
				service_name := ""
				if len(task.Spec.Networks) > 0 && len(task.Spec.Networks[0].Aliases) > 0 {
					service_name = task.Spec.Networks[0].Aliases[0]
				}
				state_current := string(task.Status.State)
				state_desired := string(task.DesiredState)
				taskStateMetric.WithLabelValues(service_name, state_current, state_desired).Set(1)
			}
			time.Sleep(time.Duration(viper.GetInt(interval)) * time.Millisecond)
		}
	}()
}

func defineParam(name, envName string, defaultValue interface{}) {
	viper.BindEnv(name, envName)
	viper.SetDefault(name, defaultValue)
}
