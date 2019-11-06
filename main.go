package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
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
		[]string{"service_name", "current_state", "desired_state", "error_details"},
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
			latestTasks := make(map[string]swarm.Task)
			list, err := cli.TaskList(ctx, types.TaskListOptions{})
			if err == nil {
				// reset metrics to delete stale values
				taskStateMetric.Reset()
				// iterate ofer task list, get latest task status
				for _, task := range list {
					taskKey := getTaskKey(task)
					if latest, ok := latestTasks[taskKey]; ok {
						if task.Status.Timestamp.After(latest.Status.Timestamp) {
							latestTasks[taskKey] = task
						}
					} else {
						latestTasks[taskKey] = task
					}
				}

				// iterate over latest tasks and gather metrics
				for taskKey, task := range latestTasks {
					state_current := string(task.Status.State)
					state_desired := string(task.DesiredState)
					error_details := task.Status.Err
					taskStateMetric.WithLabelValues(taskKey, state_current, state_desired, error_details).Set(1)
				}
			} else {
				logrus.WithFields(logrus.Fields{"error": err}).Error("Error fetching service list")
			}

			time.Sleep(time.Duration(viper.GetInt(interval)) * time.Millisecond)
		}
	}()
}

func getTaskKey(task swarm.Task) string {
	if len(task.Spec.Networks) > 0 && len(task.Spec.Networks[0].Aliases) > 0 {
		return task.Spec.Networks[0].Aliases[0] + "." + strconv.Itoa(task.Slot)
	}
	return ""
}

func defineParam(name, envName string, defaultValue interface{}) {
	viper.BindEnv(name, envName)
	viper.SetDefault(name, defaultValue)
}
