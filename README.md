# Description

Prometheus exporter for Docker Swarm task details. Currently gathers task state details (desired vs current)

# Usage

```bash
docker build -t bravo2zero/swarm-exporter .
docker run -p 8080:8080 --env METRICS_PORT=8080 --env METRICS_INTERVAL=30000  -v "/var/run/docker.sock:/var/run/docker.sock" bravo2zero/swarm-exporter
```

> localhost:8080/metrics -> swarm_task_state_details

