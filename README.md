# Alilo Agent
<img src="./images/logo.png" alt="Logo" 
     style="max-width: 300px; height: auto; display: block; margin-left: auto; margin-right: auto;">

**AliLo** Agent is an integral part of the Alilo (**Ali**express **Lo**ad) ecosystem. This tool facilitates the management of multiple concurrent K6 instances on a single machine, providing capabilities for execution control, real-time monitoring, and collection of resource utilization metrics along with script performance data.

## Scheme
<img src="./images/scheme.png" alt="Logo" 
     style="max-width: 400px; height: auto; display: block; margin-left: auto; margin-right: auto;">

The agent operates as a server-side application implementing an HTTP interface for external interactions. When processing incoming requests, the agent communicates with K6 using `exec.Cmd`. For metric collection requests related to load script status, the agent accesses k6 via its native [k6 REST API](https://grafana.com/docs/k6/latest/reference/k6-rest-api/).

### Agent endpoints

#### /api/v1/start
K6 script execution accepts script locations, test data, and runtime specifications as input parameters. The operation yields the process ID (`pid`) and comprehensive details regarding the initiated script execution.
```console
curl --location 'localhost:8888/api/v1/start' \
--header 'Content-Type: application/json' \
--data '{
    "scenarioTitle": "SomeScenarioTitile",
    "scriptTitle": "SomeScriptTitile",
    "scriptURL": "https://<your_script_location>/script.js",
    "ammoURL": "https://<your_test_data_location>/data.json",
    "params": ["-e", "RPS=2", "-e", "DURATION=20m", "-e", "STEPS=1"]
}'
```
#### /api/v1/stop
Terminates an actively executing script. Accepts the process ID (pid) to specify which instance to stop.
```console
curl --location 'localhost:8888/api/v1/stop' \
--header 'Content-Type: application/json' \
--data '{
    "pid": <pid_number_of_running_task>
}'
```
#### /api/v1/getAllTasks
Provides comprehensive information on all running script processes, including their PIDs, status, and execution metrics.
```console
curl --location --request GET 'localhost:8888/api/v1/getAllTasks' \
--header 'Content-Type: application/json' \
--data '{}'
```
#### /api/v1/getTask
Queries information about an individual running script. Accepts the process ID (`pid`) as the identifier parameter.
```console
curl --location --request GET 'localhost:8888/api/v1/getTask' \
--header 'Content-Type: application/json' \
--data '{
    "pid": <pid_number_of_running_task>
}'
```
#### /api/v1/agent/metrics
Fetches hardware resource consumption metrics for the host system where the agent is installed and operating.

```console
curl --location 'localhost:8888/api/v1/agent/metrics'
```

#### k6 REST API
K6's native [REST API](https://grafana.com/docs/k6/latest/reference/k6-rest-api/) endpoints remain accessible. The default K6 listening port will be dynamically assigned and can be retrieved from the k6ApiPort field in the JSON response of the /api/v1/start endpoint.

## Development Environment Setup
Prerequisites for modifying the agent code include having the Go programming language installed and the K6 utility downloaded on your system.

### Installing Prerequisite Packages
- golang (https://go.dev/doc/install)
- k6 (https://grafana.com/docs/k6/latest/set-up/install-k6/)

### Run agent
```console
git clone https://github.com/aliexpressru/alilo-agent.git
cd alilo-agent
go mod tidy
go run cmd/main.go
```

### Build agent
```console
go build -o alilo-agent cmd/main.go
```

### Install agent as service
For production operation, we recommend setting up the agent as a system service to ensure reliability and automatic restarts.
```console
sudo cp alilo-agent /usr/bin/alilo-agent
sudo mkdir -p /etc/alilo-agent/
sudo cp config.json /etc/alilo-agent/config.json
sudo cp alilo-agent.service /lib/systemd/system/alilo-agent.service

sudo systemctl daemon-reload
sudo systemctl enable alilo-agent.service
sudo systemctl start alilo-agent.service
sudo systemctl status alilo-agent.service
```

### System Performance Optimization
To maximize host machine performance, apply the OS-level tuning guidelines outlined in the [Grafana K6 documentation](https://grafana.com/docs/k6/latest/set-up/fine-tune-os/).
