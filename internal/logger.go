package internal

import (
	"fmt"
	"log"
	"os"

	"github.com/aliexpressru/alilo-agent/pkg/helper"
	"go.uber.org/zap"
)

var (
	atomicLevel zap.AtomicLevel

	logFileName   string
	maxLogSize    int
	maxLogBackups int
	maxLogAge     int
	logLevel      string

	logger *zap.SugaredLogger
)

func initLogger() bool {
	logFile, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Ошибка открытия файла logFileName -> '%v' ! ", logFileName)
		fmt.Println(err)
		return false
	}
	log.SetOutput(logFile)

	atomicLevel, err = helper.CreateLogger(logFileName, maxLogSize, maxLogBackups, maxLogAge, true)
	if err != nil {
		fmt.Print("Ошибка инициализации логера!")
		return false
	}
	defer func(l *zap.Logger) { //	Не уверен, что вообще эта фигня нужна. И не уверен, что она именно здесь нужна.
		fmt.Println("5")
		err = l.Sync()
		if err != nil {
			fmt.Print("Ошибка при синхронизации логера!")
			return
		}
	}(zap.L())
	logger = zap.S()
	return true
}
