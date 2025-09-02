package model

import (
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/aliexpressru/alilo-agent/pkg/utils/file_utils"
	"go.uber.org/zap"
)

var tasks = agent{
	Tasks: make(map[int]*Task),
}

type agent struct {
	sync.RWMutex
	Tasks map[int]*Task `json:"tasks"`
}

type Task struct {
	Pid            int64     `json:"pid"`
	ScenarioTitle  string    `json:"scenarioTitle"`
	ScriptTitle    string    `json:"scriptTitle"`
	ScriptURL      string    `json:"scriptURL"`
	AmmoURL        string    `json:"ammoURL"`
	ScriptFileName string    `json:"scriptFileName"`
	LogFileName    string    `json:"logFileName"`
	K6ApiPort      string    `json:"k6ApiPort"`
	PortPrometheus string    `json:"portPrometheus"`
	Params         []string  `json:"params"`
	Cmd            *exec.Cmd `json:"-"`
	LogFile        *os.File  `json:"-"`
	StartTime      time.Time `json:"startTime"`
}

func GetTask(pid int) (t *Task) {
	tasks.RLock()
	defer tasks.RUnlock()
	t, exist := tasks.Tasks[pid]
	if !exist {
		return nil
	}
	return t
}

func GetAllTasks() (ts map[int]*Task) {
	tasks.RLock()
	defer tasks.RUnlock()
	ts = tasks.Tasks
	return ts
}

func SetTask(logger *zap.SugaredLogger, pid int, t *Task) {
	tasks.Lock()
	defer tasks.Unlock()
	tasks.Tasks[pid] = t
	logger.Infof("Tasks len: %v", len(tasks.Tasks))
}

// Remove Вызов ответственен за корректную отчистку таски
func (task *Task) Remove(logger *zap.SugaredLogger) {
	//logFileName := task.LogFile.Name() //fixme: удалять старые логи как-то централизовано
	//err := os.Remove(logFileName)
	//if err != nil {
	//	logger.Warnf("Error remove file '%v'", logFileName)
	//}
	scriptFileName := task.ScriptFileName

	logger.Infof("scriptFileName: '%v'", scriptFileName)
	if scriptFileName != "" {
		err := os.Remove(scriptFileName)
		if err != nil {
			logger.Warnf("Error remove file '%v'", scriptFileName)
		}
	}
	err := task.LogFile.Close()
	if err != nil {
		logger.Warnf("Error LogFile Close:'%v'", scriptFileName)
	}
	delete(tasks.Tasks, int(task.Pid))
}

//	После ожидания duration, отправляем последние maxLogSize символов с конца лога выполняющийся команды
//
// если при вызове ожидание большое, стоит использовать горутину при вызове этой функции,
// что бы не тормозить основную логику вызова обработчика
func (task *Task) PutFootOfTaskLogInMainLog(logger *zap.SugaredLogger, duration time.Duration, maxLogSize int) string {
	defer func() {
		if er := recover(); er != nil {
			logger.Warnf("_Recovered in foot PutFootOfTaskLogInMainLog: '%v'", er)
		}
	}()

	time.Sleep(duration)
	fileName := task.LogFile.Name()
	logData := *file_utils.ReadTheData(logger, fileName)
	end := int(math.Max(float64(len(logData)-1), float64(0)))
	var start = 0
	if end > maxLogSize {
		start = end - maxLogSize
	}
	logger.Infof("Task foot log len: '%v'; start:'%v'; end:'%v'", len(logData), start, end)
	logger.Debugf("Command foot log '%v file:%v':\n'%v'", task.ScenarioTitle, fileName, logData[start:end])
	return logData[start:end]
}

//	После ожидания duration, отправляем последние maxLogSize символов c начала лога выполняющийся команды
//
// если при вызове ожидание большое, стоит использовать горутину при вызове этой функции,
// что бы не тормозить основную логику вызова обработчика
func (task *Task) PutHeadOfTaskLogInMainLog(logger *zap.SugaredLogger, duration time.Duration, maxLogSize int) string {
	defer func() {
		if er := recover(); er != nil {
			logger.Warnf("_Recovered in head PutHeadOfTaskLogInMainLog: '%v'", er)
		}
	}()

	time.Sleep(duration)
	fileName := task.LogFile.Name()
	logData := *file_utils.ReadTheData(logger, fileName)
	end := int(math.Min(float64(len(logData)-1), float64(maxLogSize)))
	var start = 0

	logger.Debugf("Task head log len: '%v'; start:'%v'; end:'%v'", len(logData), start, end)
	logger.Debugf("Command head log '%v file:%v':\n'%v'", task.ScenarioTitle, fileName, logData[start:end])
	return logData[start:end]
}

// SendTaskToLog Отправить таску в лог
func (task *Task) SendTaskToLog(logger *zap.SugaredLogger) error {
	logger.Debugf("----------task: '%+v'\n", task)
	indentTask, err := json.MarshalIndent(&task, "", "	")
	if err != nil {
		logger.Warn("Task marshaling error\n", err)
		return nil
	}
	logger.Debugf("MarshalIndent indent task: '%v'\nLogFile: '%v'", string(indentTask), task.LogFileName)
	return err
}

type PercentAgentUtilization struct {
	CpuUsed   int `json:"cpuUsed"`
	MemUsed   int `json:"memUsed"`
	PortsUsed int `json:"portsUsed"`
}
