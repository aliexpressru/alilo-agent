package time_utils

import (
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
)

func TimeOut(logger *zap.SugaredLogger, delayToServices string, start time.Time) {
	var (
		duration time.Duration
	)

	diff := time.Now().UnixNano() - start.UnixNano()
	logger.Infof("diff 		%v", diff)
	worked := strconv.FormatInt(diff, 10)
	logger.Infof("worked 		%v", worked)
	duration2, err := time.ParseDuration(fmt.Sprint(worked, "ns"))
	if err != nil {
		logger.Warn("Косяк в указании времени ожидания сервиса - 1")
		return
	}
	logger.Infof("duration2	%v", duration2)

	duration, err = time.ParseDuration(delayToServices)
	if err != nil {
		logger.Warn("Косяк в указании времени ожидания сервиса - 2")
		return
	}
	logger.Infof("duration	%v", duration)

	sum := duration.Nanoseconds() - duration2.Nanoseconds()
	newDuration, err := time.ParseDuration(fmt.Sprint(sum, "ns"))
	if err != nil {
		logger.Warn("Косяк в указании времени ожидания сервиса - 3")

		return
	}
	newDuration = newDuration.Round(time.Nanosecond)
	logger.Infof("newDuration	%v", newDuration)

	logger.Infof("Ждем %v", newDuration)
	time.Sleep(newDuration)
	logger.Infof("Дождались %v", newDuration)
}

// Вернуть кол-во секунд int, в виде сетруктуры Duration
func GetDurationInSec(numberOfSec int) time.Duration {
	return GetDuration(numberOfSec, time.Second)
}

// Вернуть кол-во времени в виде указанной сетруктуры Duration
func GetDuration(numberOfSec int, duration time.Duration) time.Duration {
	return time.Duration(numberOfSec) * duration
}
