package utils

import (
	"net"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap"
)

func FindOutTheIpAddress(logger *zap.SugaredLogger) (ip string) {
	hostname, _ := os.Hostname()
	logger.Info("Hostname: ", hostname)

	addrs, _ := net.InterfaceAddrs()
	logger.Debug("addrs: ", addrs)
	logger.Debug("OS: ", os.Getenv("OS"))

	if strings.Contains(os.Getenv("OS"), "Windows") {
		addr := addrs[0].String() // берем первый для ОС Win
		logger.Debug("addrs[0]: ", addr)
		ip = strings.Split(addr, "/")[0]
		logger.Debug("Windows ip: ", ip)
	} else if strings.Contains(runtime.GOOS, "darwin") {
		logger.Debug("addrs len: ", len(addrs))
		addr := addrs[len(addrs)-1].String() // берем последний для ОС Linux
		logger.Debug("addrs[12]: ", addr)
		ip = strings.Split(addr, "/")[0]
		logger.Debug("ip: ", ip)
	} else {
		addr := addrs[1].String() // берем первый для ОС Linux
		logger.Debug("addrs[1]: ", addr)
		ip = strings.Split(addr, "/")[0]
		logger.Debug("ip: ", ip)
	}
	return ip
}
