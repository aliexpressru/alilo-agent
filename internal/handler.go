package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aliexpressru/alilo-agent/internal/model"
	"github.com/aliexpressru/alilo-agent/pkg/helper"
	"github.com/aliexpressru/alilo-agent/pkg/utils/file_utils"
	"github.com/aliexpressru/alilo-agent/pkg/utils/http_utils"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const (
	ResponseStatusError  = "Error"
	ResponseStatusStatus = "Success"

	ResponseErrorDefunct          = "the process is defunct"
	ResponseErrorNoSuchTask       = "there is no such task"
	ResponseErrorLogFileNotExist  = "log file is not exist"
	ResponseErrorNoPortsAvailable = "no ports available"
)

func GetAllTasks(w http.ResponseWriter, r *http.Request) {
	http_utils.Foreword(logger, r)
	allTasks := model.GetAllTasks()
	var respGetAllTasks = model.ResponseGetAllTasks{Tasks: allTasks}
	returnResponse(w, &respGetAllTasks, ResponseStatusStatus)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	rqData := http_utils.Foreword(logger, r)
	var rq = model.Request{}
	var respGetAllTasks = model.ResponseGetAllTasks{}
	if rqData == "" {
		returnErrorResponse(w, &respGetAllTasks, http.ErrBodyNotAllowed)
		return
	}
	if err := json.Unmarshal([]byte(rqData), &rq); err != nil {
		returnErrorResponse(w, &respGetAllTasks, err)
		return
	}
	logger.Info("GetTask received process pid: ", rq.Pid)
	respGetAllTasks.Tasks = map[int]*model.Task{
		rq.Pid: model.GetTask(rq.Pid),
	}
	returnResponse(w, &respGetAllTasks, ResponseStatusStatus)
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	rqData := http_utils.Foreword(logger, r)
	var rq = model.Request{}
	var respGetStatus = model.ResponseGetStatus{}
	if rqData == "" {
		logger.Warnf("Body: '%v' Error: '%v'", rqData, http.ErrBodyNotAllowed.Error())
		returnErrorResponse(w, &respGetStatus, http.ErrBodyNotAllowed)
		return
	}
	if err := json.Unmarshal([]byte(rqData), &rq); err != nil {
		logger.Warnf("Unmarshal error: '%v' \nrqData'%v'", err, rqData)
		returnErrorResponse(w, &respGetStatus, err)
		return
	}

	logger.Info("GetStatus received process pid: ", rq.Pid)
	task := model.GetTask(rq.Pid)
	if task == nil {
		err := errors.New(ResponseErrorNoSuchTask)
		logger.Warnf("Task pid: '%+v' -> error:'%+v'", rq.Pid, err)
		returnErrorResponse(w, &respGetStatus, err)
		return
	}
	var err error
	var wg = &sync.WaitGroup{}
	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		defer func(wg *sync.WaitGroup) {
			wg.Done()
		}(wg)
		respGetStatus.Task = task
		task.PutFootOfTaskLogInMainLog(logger, time.Nanosecond, 4000)
		err = healthCheck(strconv.FormatInt(task.Pid, 10))
	}(wg)
	go func(wg *sync.WaitGroup) {
		defer func(wg *sync.WaitGroup) {
			wg.Done()
		}(wg)
		metricsResponse, er := getScriptLoadMetrics(task.K6ApiPort)
		if er != nil {
			logger.Warnf("getScriptLoadMetrics error! PID: '%+v' ERROR: '%+v' metricsResponse: '%+v'", rq.Pid, er, metricsResponse)
		} else {
			logger.Debugf("getScriptLoadMetrics PID: '%+v' metricsResponse: '%+v'", rq.Pid, metricsResponse)
		}
		respGetStatus.Metrics = pullAScriptMetricStructure(metricsResponse)
	}(wg)

	wg.Wait()
	if err != nil {
		logger.Warnf("healthCheck PID: '%+v' ERROR: '%+v'", rq.Pid, err)
		wg.Add(2)
		go func(wg *sync.WaitGroup) {
			defer func(wg *sync.WaitGroup) {
				wg.Done()
			}(wg)
			errUnlockPortK6 := helper.NewUnlockPort(task.K6ApiPort)
			if errUnlockPortK6 != nil {
				logger.Warn("GetStatus UnlockPort err: ", errUnlockPortK6)
			}
		}(wg)
		go func(wg *sync.WaitGroup) {
			defer func(wg *sync.WaitGroup) {
				wg.Done()
			}(wg)
			errUnlockPortPrometheus := helper.NewUnlockPort(task.PortPrometheus)
			if errUnlockPortPrometheus != nil {
				logger.Warn("GetStatus UnlockPort err: ", errUnlockPortPrometheus)
			}
		}(wg)
		wg.Wait()
		returnErrorResponse(w, &respGetStatus, err)
		return
	}
	returnResponse(w, &respGetStatus, ResponseStatusStatus)
}

func GetTaskLogs(w http.ResponseWriter, r *http.Request) {
	http_utils.Foreword(logger, r)
	var rq = model.RequestGetTaskLogs{}
	rq.Name = r.URL.Query().Get("name")
	rq.Pid, _ = strconv.Atoi(r.URL.Query().Get("pid"))
	rq.Len, _ = strconv.Atoi(r.URL.Query().Get("len"))
	rq.Head, _ = strconv.ParseBool(r.URL.Query().Get("head"))
	w.Header().Set("Content-Type", "text/plain;charset=UTF-8")

	logger.Infof("GetTaskLogs received process Pid:'%v' Len:'%v' Name:'%v'  Head:'%v' r.URL.Query():'%v'",
		rq.Pid, rq.Len, rq.Name, rq.Head, r.URL.Query())
	if rq.Len == 0 {
		rq.Len = 4000
	}
	var logData = ""
	if rq.Pid != 0 {
		task := model.GetTask(rq.Pid)
		if task != nil {
			if rq.Head {
				logData = task.PutHeadOfTaskLogInMainLog(logger, time.Nanosecond, rq.Len)
			} else {
				logData = task.PutFootOfTaskLogInMainLog(logger, time.Nanosecond, rq.Len)
			}
		} else {
			logger.Warnf("RqBody: '%v' Error: '%v'", rq, http.ErrBodyNotAllowed.Error())
			http.Error(w,
				fmt.Sprintf(
					"Bad request. The active process(Task) log request is waiting for an pid, Pid '%v' has already been completed or has not been started..",
					rq.Pid),
				400)
			return
		}
	} else if rq.Name != "" {
		var start, end int
		logData = *file_utils.ReadTheData(logger, rq.Name) // fixme: это пипец как не оптимально, если скрипт будет работать долго - лог будет огромный, нет смысла его весь читать

		if rq.Head { // Как будем читать лог, с начала или с конца файла
			end = int(math.Min(float64(len(logData)-1), float64(rq.Len))) // fixme: нужно вынести формулу вычисления в фунци
		} else {
			end = int(math.Max(float64(len(logData)-1), float64(0)))
			if end > rq.Len {
				start = end - rq.Len
			}
		}
		logger.Infof("Task log{ len: '%v'; start:'%v'; end:'%v'}", len(logData), start, end)
		logData = logData[start:end]
	}

	if logData != "" {
		_, err := fmt.Fprintf(w, "%s", logData) //возвращаем ответ на запрос
		if err != nil {
			logger.Warn(err)
		}
	} else {
		logger.Warnf("RqBody: '%v' Error: '%v'", rq, http.ErrBodyNotAllowed.Error())
		http.Error(w, "Bad request. The use of the 'name', 'pid', 'head' and 'len' parameters is required.", 400)
	}
}

func SaveScript(w http.ResponseWriter, r *http.Request) {
	rqData := http_utils.Foreword(logger, r)
	var respSaveScript model.ResponseSaveScript
	if rqData != "" {
		tFile := saveScriptToFile(rqData, "SaveScript")

		respSaveScript = model.ResponseSaveScript{PathScript: tFile.Name()}
		returnResponse(w, &respSaveScript, ResponseStatusStatus)
		return
	}
	returnErrorResponse(w, &respSaveScript, http.ErrBodyNotAllowed)
}

func StopCommand(w http.ResponseWriter, r *http.Request) {
	rqData := http_utils.Foreword(logger, r)
	respStopCommand := model.Response{}
	//nolint
	status := ResponseStatusError
	var rq = model.Request{}
	if err := json.Unmarshal([]byte(rqData), &rq); err == nil {
		respStopCommand.Pid = rq.Pid
		logger.Info("Received process pid to stop: ", rq.Pid)
		if task := model.GetTask(rq.Pid); task != nil {
			cmd := task.Cmd
			//helper.CleanUpPorts(logger, task.LogFileName)
			// TODO: освобождение порта
			errUnlockPortK6 := helper.NewUnlockPort(task.K6ApiPort)
			if errUnlockPortK6 != nil {
				logger.Warn("StopCommand UnlockPort err: ", errUnlockPortK6)
			}
			errUnlockPortPrometheus := helper.NewUnlockPort(task.PortPrometheus)
			if errUnlockPortPrometheus != nil {
				logger.Warn("StopCommand UnlockPort err: ", errUnlockPortPrometheus)
			}
			if cmd == nil {
				logger.Warnf("There is no such cmd Pid:'%v' CMD:'%v' ", rq.Pid, cmd)
				returnErrorResponse(w, &respStopCommand, errors.New(fmt.Sprint("there is no such Cmd in task ", strconv.Itoa(rq.Pid))))
				return
			}
			logger.Debug("healthCheck started: ", cmd)
			err = healthCheck(strconv.FormatInt(task.Pid, 10))
			if err != nil {
				logger.Warnf("Stop task healthCheck: '%+v'", err)
			}
			logger.Debug("Command to stop: ", cmd)
			if err = cmd.Process.Signal(syscall.Signal(2)); err != nil {
				logger.Warnf("Task '%v' is not stopped!", rq.Pid)
				logger.Warn("Kill error: ", err)
				returnErrorResponse(w, &respStopCommand, err)
				return
			}
			var state *os.ProcessState
			state, err = cmd.Process.Wait()
			if err != nil {
				logger.Warn("Wait error: ", err)
				returnErrorResponse(w, &respStopCommand, err)
				return
			}
			logger.Debug("State: ", state)
			status = state.String()
			go task.PutFootOfTaskLogInMainLog(logger, time.Nanosecond, 4000)
			task.Remove(logger)
		} else {
			logger.Warn("There is no such task ", rq.Pid)
			returnErrorResponse(w, &respStopCommand, errors.New(ResponseErrorNoSuchTask))
			return
		}
		respStopCommand.Pid = rq.Pid
		returnResponse(w, &respStopCommand, status)
	} else {
		logger.Warnf("Stopped anmarshaling error\n '%v'", err)
		returnErrorResponse(w, &respStopCommand, err)
		return
	}
}

func StartCommand(w http.ResponseWriter, r *http.Request) {
	rqData := http_utils.Foreword(logger, r)
	respStartCommand := model.Response{}
	logger.Infof("----------rqData '%v'", rqData)
	if rqData == "" {
		logger.Warnf("StartCommand the body is empty -> '%v' \n", rqData)
		returnErrorResponse(w, &respStartCommand, http.ErrBodyNotAllowed)
		return
	}
	task, err := prepareCommandTask(rqData)
	if err != nil {
		returnErrorResponse(w, &respStartCommand, err)
		return
	}

	scriptRunCmd := prepareCommandToStart(task)
	if scriptRunCmd == nil {
		message := fmt.Sprintf("Prepare command to start err:  '%+v'", scriptRunCmd)
		logger.Debug(message)
		err = errors.New(message)
		returnErrorResponse(w, &respStartCommand, err)
		return
	}
	err = scriptRunCmd.Start()
	if err != nil {
		logger.Warnf("--- Start connand err:  '%+v'", err)
		if task.LogFile != nil && file_utils.FileExist(task.LogFileName) {
			respStartCommand.Error = *file_utils.ReadTheData(logger, task.LogFile.Name())
		} else {
			respStartCommand.Error = ResponseErrorLogFileNotExist
		}
		respStartCommand.Pid = scriptRunCmd.Process.Pid
		returnErrorResponse(w, &respStartCommand, err)
		return
	}
	logger.Debugf("----------scriptRunCmd.Process Process.Pid:'%+v' Process:'%+v'",
		scriptRunCmd.Process.Pid, scriptRunCmd.Process)
	task.Cmd = scriptRunCmd
	task.Pid = int64(scriptRunCmd.Process.Pid)
	model.SetTask(logger, scriptRunCmd.Process.Pid, task)
	go task.PutFootOfTaskLogInMainLog(logger, time.Second*10, 2000)
	if file_utils.FileExist(task.LogFile.Name()) {
		respStartCommand.Error = *file_utils.ReadTheData(logger, task.LogFile.Name())
	} else {
		respStartCommand.Error = ResponseErrorLogFileNotExist
	}
	respStartCommand.Pid = scriptRunCmd.Process.Pid
	respStartCommand.Task = task
	returnResponse(w, &respStartCommand, ResponseStatusStatus)
}

func GetMetricsAgentResourceUtilization(w http.ResponseWriter, r *http.Request) {
	respUtilization := model.ResponseUtilization{}
	agentUtilization, err := getMetricsAgentResourceUtilization()
	if err != nil {
		fmt.Printf("Не удалось получить данные об утилизации: %s", err)
		respUtilization.PercentAgentUtilization = agentUtilization
		returnErrorResponse(w, &respUtilization, err)
		return
	}
	respUtilization.PercentAgentUtilization = agentUtilization
	returnResponse(w, &respUtilization, ResponseStatusStatus)
}

func getMetricsAgentResourceUtilization() (metrics model.PercentAgentUtilization, err error) {
	var cpuPercent []float64
	var errPercent error
	var vmStat *mem.VirtualMemoryStat
	var errMemory error
	var percentPortsUsed float64

	cpuPercent, errPercent = cpu.Percent(time.Second, false)
	if errPercent != nil {
		fmt.Printf("Не удалось получить данные об утилизации ЦПУ: %s\n", errPercent)
	}
	logger.Infof("Процент утилизации ЦПУ: %.2f%%", cpuPercent[0])

	vmStat, errMemory = mem.VirtualMemory()
	if errMemory != nil {
		fmt.Printf("Не удалось получить данные об утилизации ОЗУ: %s\n", errMemory)
	}
	logger.Infof("Процент утилизации ОЗУ: %.2f%%", vmStat.UsedPercent)

	percentPortsUsed = helper.NewPercentPortsUsed()
	logger.Infof("Процент занятых портов: %.2f%%", percentPortsUsed)

	return model.PercentAgentUtilization{
		CpuUsed:   int(cpuPercent[0]),
		MemUsed:   int(vmStat.UsedPercent),
		PortsUsed: int(percentPortsUsed),
	}, errors.Join(errPercent, errMemory)
}

func saveScriptToFile(rqData string, title string) *os.File {
	return file_utils.SaveDataScriptToFile(logger, rqData, title)
}

func saveScriptFromURLToFile(uri string, scenarioTitle string, scriptTitle string) (*os.File, bool) {
	result := false
	defer func() {
		if er := recover(); er != nil {
			logger.Warnf("Recovered in saveScriptFromURLToFile. error: '%v'", er)
			result = false
		}
	}()
	if uri == "" { //todo: +проверять на валидность адреса
		logger.Warnf("----Start saved script 1")
		return nil, result
	}
	logger.Infof("----Start saved script 2")
	data := http_utils.GetWithHeaders(logger, uri, map[string]string{})
	logger.Infof("-----------Get(Script file data '%.50v')", data)
	if data == "" && strings.Contains(data, "<Message>The specified key does not exist.</Message>") {
		logger.Warn("Script file data == '', file noe exist or empty")
		return nil, result
	} else if data == "" && strings.Contains(data, "404 File Not Found") {
		logger.Warn("Script file data == '', file noe exist or empty")
		return nil, result
	} else if data == "" && strings.Contains(data, "401 Unauthorized") {
		logger.Warn("Script file data == '', Invalid token")
		return nil, result
	} else if data == "" && strings.Contains(data, "404 Commit Not Found") {
		logger.Warn("Script file data == '', Invalid ref param in URL")
		return nil, result
	} else if data == "" && strings.Contains(data, "404 Not Found") {
		logger.Warn("Script file data == '', Invalid URL")
		return nil, result
	} else {
		result = true
	}
	// подставляем название сценария в скрипт:
	if strings.Contains(data, "alilo_load: {") {
		//scenario = string_utils.ReplaceAllUnnecessarySymbols(scenario)
		logger.Warnf("Updating the scenario title in the options in master script: '%v'", scenarioTitle)
		data = strings.Replace(data, "alilo_load: {", fmt.Sprint(scenarioTitle, ": {"), 1)
	}

	return file_utils.SaveDataScriptToFile(logger, data, scriptTitle), result
}

func saveAmmoFromURLToFile(uri string) (*os.File, bool) {
	fileName := path.Base(uri)
	result := false
	defer func() {
		if er := recover(); er != nil {
			logger.Warnf("Recovered in saveScriptFromURLToFile. error: '%v'", er)
			result = false
		}
	}()
	if uri == "" { //todo: +проверять на валидность адреса
		logger.Warnf("----Start saved ammo 1")
		return nil, result
	}
	logger.Infof("----Start saved ammo 2")
	data := http_utils.GetWithHeaders(logger, uri, map[string]string{})
	logger.Infof("-----------Get(Ammo file data '%.50v')", data)
	if data == "" && strings.Contains(data, "<Message>The specified key does not exist.</Message>") {
		logger.Warn("Ammo file data == '', file noe exist or empty")
		return nil, result
	} else if data == "" && strings.Contains(data, "404 File Not Found") {
		logger.Warn("Ammo file data == '', file noe exist or empty")
		return nil, result
	} else if data == "" && strings.Contains(data, "401 Unauthorized") {
		logger.Warn("Ammo file data == '', Invalid token")
		return nil, result
	} else if data == "" && strings.Contains(data, "404 Commit Not Found") {
		logger.Warn("Ammo file data == '', Invalid ref param in URL")
		return nil, result
	} else if data == "" && strings.Contains(data, "404 Not Found") {
		logger.Warn("Ammo file data == '', Invalid URL")
		return nil, result
	} else {
		result = true
	}
	return file_utils.SaveDataAmmoToFile(logger, data, fileName), result
}

func healthCheck(pid string) error {
	cmd := exec.Command("ps",
		"-fp", pid,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var out, outErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &outErr
	logger.Info("healthCheck cmd: ", cmd.String())
	err := cmd.Start()
	if err != nil {
		logger.Warnf("healthCheck Start err:  '%v'", err)
		return err
	}

	state, err := cmd.Process.Wait()
	if err != nil {
		logger.Warn("healthCheck Wait error: ", err)
		return err
	}
	time.Sleep(100 * time.Millisecond)
	outData := fmt.Sprintf("State:'%v'\nStdout:'%v'\nStderr:'%v'\n",
		state.String(), out.String(), outErr.String())
	logger.Debugf("healthCheck: '%v'\n", outData)
	logger.Debugf("healthCheck State: %v", state)
	if state.ExitCode() != 0 {
		err = errors.New(fmt.Sprint("healthCheck state: ", outData))
		return err
	}
	return processingOutputPSCommand(out.String(), pid)
}

func processingOutputPSCommand(data string, pid string) error {
	if !strings.Contains(data, pid) {
		message := "the process is not running"
		logger.Warnf("processingOutputPSCommand: %v", message)
		return errors.New(message)
	}
	logger.Info("Task exist: ", pid)
	split := strings.Split(data, "\n")
	logger.Debug("Output line count: ", len(split))
	for _, line := range split {
		if strings.Contains(line, pid) {
			logger.Debugf("PS output:\n'%v'\n", line)
			if strings.Contains(line, "<defunct>") {
				message := ResponseErrorDefunct
				logger.Infof("processingOutputPSCommand: %v", message)
				return errors.New(message)
			}
		}
	}
	return nil
}

func prepareCommandToStart(task *model.Task) *exec.Cmd {
	scriptFile, done := saveScriptFromURLToFile(task.ScriptURL, task.ScenarioTitle, task.ScriptTitle)
	if !done {
		logger.Warnf("script file saving error")
		return nil
	}

	ammoFile, done := saveAmmoFromURLToFile(task.AmmoURL)
	if !done {
		logger.Warnf("ammo file saving error")
		return nil
	}
	logger.Infof("----------scriptFile '%.100v'", scriptFile.Name())
	task.ScriptFileName = scriptFile.Name()
	params := prepareParams(task.ScriptFileName, task.Params, task.K6ApiPort, task.PortPrometheus)
	params = append(params, "-e")
	params = append(params, fmt.Sprintf("AMMO_URL=%s", ammoFile.Name()))

	scriptRunCmd := exec.Command("k6",
		params...,
	)
	scriptRunCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	scriptRunCmd.Stdout = task.LogFile
	scriptRunCmd.Stderr = task.LogFile
	logger.Warnf("---------- Test run command: '%v'", scriptRunCmd.String())

	task.AmmoURL = ammoFile.Name()
	return scriptRunCmd
}

func prepareCommandTask(rqData string) (task *model.Task, err error) {
	logger.Debugf("prepareCommandTask rqData: %v", rqData)
	task = &model.Task{}
	if err = json.Unmarshal([]byte(rqData), &task); err != nil {
		logger.Warnf("Task anmarshaling error\n '%v'", err)
		return nil, err
	}
	task.StartTime = time.Now()
	var logFile *os.File
	var title = task.ScenarioTitle
	if task.ScriptTitle != "" {
		title = task.ScriptTitle
	}
	logFile, err = file_utils.GetNewFile(logger, fmt.Sprint("TaskLog_", strings.ReplaceAll(title, " ", "_")))
	if err != nil {
		logger.Warnf("Task GetNew logFile error:\n'%v'", err)
		return nil, err
	}
	task.LogFile = logFile
	task.LogFileName = logFile.Name()
	task.K6ApiPort, err = helper.NewGetFreePort(logger)
	if err != nil {
		logger.Warnf("Task GetFreePort error:'%v'", err)
		if strings.Contains(err.Error(), "ports have run out") {
			return nil, fmt.Errorf(ResponseErrorNoPortsAvailable)
		}
		return nil, err
	}
	task.PortPrometheus, err = helper.NewGetFreePort(logger)
	if err != nil {
		logger.Warnf("Task GetFreePort error:'%v'", err)
		return nil, err
	}
	logger.Debugf("prepareCommandTask task: %v", task)
	err = task.SendTaskToLog(logger)
	if err != nil {
		logger.Warnf("Task SendTaskToLog error:\n'%v'", err)
		return nil, err
	}
	return task, err
}

func prepareParams(scriptFileName string, taskParams []string, k6ApiPort string, portPrometheus string) []string {
	params := []string{
		"run",
		scriptFileName,
		"--address", fmt.Sprint("0.0.0.0:", k6ApiPort),
	}
	for i, param := range taskParams {
		if param == "prometheus=port=" {
			taskParams[i] = fmt.Sprint(taskParams[i], portPrometheus)
		}
	}
	params = append(params, taskParams...)
	logger.Infof("----------params: '%v'", params)
	return params
}

func getScriptLoadMetrics(port string) (rs model.MetricsResponse, err error) {
	logger.Infof("getScriptLoadMetrics port: %v", port)
	url := fmt.Sprintf("%v:%v%v", "http://0.0.0.0", port, "/v1/metrics")
	defer func() {
		if er := recover(); er != nil {
			logger.Warnf("getScriptLoadMetrics failed:{error:{%+v}, url:{%v}}", er, url)
		}
	}()
	//	curl -X GET http://10.41.240.200:6652/v1/metrics -H 'Content-Type: application/json'
	rsBody, err := http_utils.Get(logger, url)
	if err != nil {
		return rs, errors.Join(err, fmt.Errorf("metrics Get request error"))
	}
	if rsBody == nil {
		return rs, errors.Join(err, errors.New("metrics is nil"))
	}
	logger.Debugf("getScriptLoadMetrics rsBody: %s", rsBody)
	if err != nil {
		err = errors.Join(err, fmt.Errorf("failed call getScriptLoadMetrics '%v'", url))
		logger.Warnf("getScriptLoadMetrics: %v", err.Error())
		return rs, err
	}
	logger.Debugf("getScriptLoadMetrics 1 (RsBody:'%v', Response struct:'%+v')", string(rsBody), rs)
	err = json.Unmarshal(rsBody, &rs)
	if err != nil {
		err = errors.Join(err, errors.New("unmarshal error"))
		logger.Warn(err)
		return rs, err
	}
	logger.Debugf("getScriptLoadMetrics 2 (RsBody:'%v', Response struct:'%+v')", string(rsBody), rs)
	return rs, err
}

func pullAScriptMetricStructure(metrs model.MetricsResponse) *model.Metrics {
	m := model.Metrics{
		Rps:      "0",
		Rt90P:    "0",
		Rt95P:    "0",
		Rt99P:    "0",
		RtMax:    "0",
		Failed:   "0",
		Vus:      "0",
		Sent:     "0",
		Received: "0",
	}
	logger.Infof("Metrics incoming: %v", metrs)
	for _, entity := range metrs.Data {
		if entity.ID == model.K6HttpReqs {
			m.Rps = fmt.Sprintf("%.1f", entity.Attributes.Sample.Rate)
			logger.Infof("setting the Rps metric: %v", m.Rps)
		}
		if entity.ID == model.K6HttpReqsFailed {
			m.Failed = fmt.Sprintf("%.1f", entity.Attributes.Sample.Rate)
			logger.Infof("setting the Failed metric: %v", m.Failed)
		}
		if entity.ID == model.K6Vus {
			m.Vus = fmt.Sprintf("%v", entity.Attributes.Sample.Value)
			logger.Infof("setting the Vus metric: %v", m.Vus)
		}
		if entity.ID == model.K6DataSent {
			m.Sent = fmt.Sprintf("%v", entity.Attributes.Sample.Count)
			logger.Infof("setting the Sent metric: %v", m.Sent)
		}
		if entity.ID == model.K6DataReceived {
			m.Received = fmt.Sprintf("%v", entity.Attributes.Sample.Count)
			logger.Infof("setting the Received metric: %v", m.Received)
		}
		if entity.ID == model.K6HttpReqDuration {
			logger.Infof("setting the Duration metric: %v", m.Rt90P)
			m.Rt90P = fmt.Sprintf("%.1f", entity.Attributes.Sample.P90)
			m.Rt95P = fmt.Sprintf("%.1f", entity.Attributes.Sample.P95)
			m.RtMax = fmt.Sprintf("%.1f", entity.Attributes.Sample.Max)
		}
	}
	logger.Infof("Metrics outgoing: %v", m)
	return &m
}
