package helper

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/aliexpressru/alilo-agent/pkg/utils/string_utils"
	"go.uber.org/zap"
)

var (
	startPort = 6565
	CountPort = 600
	//maxPort   = startPort + CountPort
	//ports     = &activePorts{mActivePorts: make(map[int]string)}

	newPorts *newActivePorts
)

//type activePorts struct {
//	mu sync.RWMutex
////	map[port]logFileName - logFileName т.к. на момент получения порта PID у процесса еще не выделен, по этому имя лога остается единственным уникальным атрибутом
//mActivePorts map[int]string
//}

//func appendPort(port int, logFileName string) {
//	ports.mu.Lock()
//	defer ports.mu.Unlock()
//	ports.mActivePorts[port] = logFileName
//}
//
//func existPort(port int) bool {
//	ports.mu.RLock()
//	defer ports.mu.RUnlock()
//	_, ok := ports.mActivePorts[port]
//	return ok
//}
//
//func CleanUpPorts(logger *zap.SugaredLogger, logFileName string) {
//	defer func() {
//		if er := recover(); er != nil {
//			logger.Warnf("_Recovered in CleanUpPorts: '%v' '%v'", logFileName, er)
//		}
//	}()
//
//	ports.mu.Lock()
//	defer ports.mu.Unlock()
//	logger.Infof("Strat CleanUpPorts %v", logFileName)
//	for aPort, aLogFileName := range ports.mActivePorts {
//		logger.Debugf("CleanUpPort range: \n%v\n%v", aLogFileName, logFileName)
//		if logFileName == aLogFileName {
//			logger.Infof("CleanUpPort aPort: %v", aPort)
//			delete(ports.mActivePorts, aPort)
//		} else {
//			logger.Debugf("CleanUpPort port missed: %v", aPort)
//		}
//	}
//	logger.Infof("Stop CleanUpPorts %v", logFileName)
//}

type newActivePorts struct {
	// map[port]Mutex - используемый порт всегда будет заблокирован на запись
	sync.RWMutex
	mActivePorts map[int]*sync.RWMutex
}

// func NewinitActivePorts(startPort int) {
func init() {
	newPorts = &newActivePorts{mActivePorts: make(map[int]*sync.RWMutex, CountPort)}
	curPort := startPort
	for i := 0; i < CountPort; curPort++ { // в конце каждого цикла увеличиваем порт
		if !PortIsOpened("0.0.0.0", curPort) {
			newPorts.mActivePorts[curPort] = &sync.RWMutex{}
			// если порт свободный, добавляем его и увеличиваем счетчик(что-бы в результате е нас было ровно CountPort в пачке)
			i++
		}
	}
}

func NewCountPortsUsed() (countPort int) {
	newPorts.Lock()
	defer newPorts.Unlock()
	for p, mutex := range newPorts.mActivePorts {
		if mutex.TryLock() {
			mutex.Unlock()
		} else {
			countPort++
			zap.S().Infof("- %v порт заблокирован", p)
		}
	}
	return countPort
}

func NewPercentPortsUsed() (countPort float64) {
	availablePorts := NewCountPortsUsed()
	zap.S().Infof("Количество занятых портов: %v", availablePorts)
	return math.Ceil(float64(availablePorts) / float64(CountPort) * 100)
}

func newGetFreePort() (port int, err error) {
	newPorts.Lock()
	defer newPorts.Unlock()
	for p, mutex := range newPorts.mActivePorts {
		if mutex.TryLock() {
			if PortIsOpened("0.0.0.0", p) {
				zap.S().Warnf("GetFreePort '%v' TryLock true, address already in use", p)
				mutex.Unlock()
			}
			return p, nil
		} else {
			zap.S().Infof("GetFreePort '%v' TryLock false", p)
		}
	}

	return port, fmt.Errorf("ports have run out")
}

func NewUnlockPort(port string) error {
	newPorts.Lock()
	defer newPorts.Unlock()
	intPort, err := string_utils.StrToInt(port)
	if err != nil {
		zap.S().Warn("UnlockPort '%v' StrToInt err: ", port, err)
	}
	mutex, exist := newPorts.mActivePorts[intPort]
	if exist {
		mutex.TryLock()
		mutex.Unlock()
	} else {
		return fmt.Errorf("the port is not in use")
	}
	return nil
}

func PortIsOpened(host string, port int) (res bool) {
	timeout := time.Second
	target := fmt.Sprintf("%v:%v", host, port)
	conn, err := net.DialTimeout("tcp", target, timeout)
	//target := fmt.Sprintf(":%v", port)
	//conn, err := net.Listen("tcp", target) // fixme: заменить DialTimeout на Listen
	if err != nil {
		return res
	}
	defer func() {
		if conn != nil {
			err = conn.Close()
			if err == nil {
				res = true
			}
		}
	}()

	return res
}

// NewGetFreePort return free port for the used in running k6 scripts
// 6565 for --address "0.0.0.0:6565"
func NewGetFreePort(logger *zap.SugaredLogger) (addres string, err error) {
	logger.Infof("GetFreePort")
	freePort, err := newGetFreePort()
	if err != nil {
		logger.Warnf("GetFreePort error: %v", err)
		return "", err
	}
	logger.Infof("Free port for K6 API: '%v'", freePort)
	return strconv.Itoa(freePort), nil
}
