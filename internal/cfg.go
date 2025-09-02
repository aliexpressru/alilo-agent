package internal

import (
	"bytes"
	"encoding/json"

	"github.com/aliexpressru/alilo-agent/internal/model"
	"github.com/aliexpressru/alilo-agent/pkg/helper"
	"github.com/aliexpressru/alilo-agent/pkg/utils/file_utils"
	"go.uber.org/zap"
)

var (
	cfg            model.Config
	configFileName string
)

func initCfg() bool {
	if fileData, err := file_utils.ReadBytesFromFile(logger, configFileName); err == nil {
		// Устанавливаем обязательные значения по умолчанию
		cfg = model.Config{
			ServerPort: serverPort,
			LogLevel:   logLevel,
		}

		trimFileData := bytes.Trim(fileData, file_utils.BOM) // work only without BOM
		if err = json.Unmarshal(trimFileData, &cfg); err == nil {
			helper.SetLoggerLevel(atomicLevel, cfg)
			cfgToLog := cfg
			marshalCfg, errMarshal := json.MarshalIndent(&cfgToLog, "", "	")
			if errMarshal != nil {
				logger.Error("Ошибка преобразования конфига", errMarshal)
				return false
			} else {
				logger.Debug("MarshalIndent cfg: ", string(marshalCfg))
			}
		} else {
			logger.Panicf("Ошибка анмаршалинга конфигурационного файла %v", zap.Error(err))
			return false
		}
	} else {
		logger.Panicf("Ошибка чтения файла configFileName -> '%v' \nerror: '%v'\n", configFileName, err)
		return false
	}
	return true
}
