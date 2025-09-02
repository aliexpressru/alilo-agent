package file_utils

import (
	"container/list"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aliexpressru/alilo-agent/pkg/utils/string_utils"
	"go.uber.org/zap"
)

const BOM = "\xef\xbb\xbf"

// ReadTheData - Получаем данные из файла
func ReadTheData(logger *zap.SugaredLogger, filePath string) (fileData *string) {
	if file, err := ReadBytesFromFile(logger, filePath); err == nil {
		return string_utils.ReturnPointer(string(file))
	} else {
		logger.Warnf("ReadBytesFromFile err:%v\n", err)
	}
	return string_utils.ReturnPointer("")
}

// ReadBytesFromFile - Получаем байты из файла
func ReadBytesFromFile(logger *zap.SugaredLogger, filePath string) (fileData []byte, err error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		logger.Info("ReadBytesFromFile err:\n", err)
	}
	return file, err
}

// GetListFiles - Получаем список доступных файлов в указанной дирректории
func GetListFiles(logger *zap.SugaredLogger, dirname string) (listFiles list.List) {
	logger.Infof("Processing create servises from -> '%v'", dirname)

	// если совсем плохо с путем, меняем его на getwd
	if len(dirname) < 1 {
		dirname, _ = os.Getwd()
	}

	if fileInfo, err := os.ReadDir(dirname); err == nil {
		for fileNum := range fileInfo {
			logger.Infof("file name: -> %s", fileInfo[fileNum].Name())
			filePath := filepath.Join(dirname, fileInfo[fileNum].Name())
			if !fileInfo[fileNum].IsDir() {
				logger.Infof("Is file: %v", filePath)
				listFiles.PushBack(filePath)
			} else {
				logger.Infof("Is dir: %v", filePath)
				reader := GetListFiles(logger, filePath)
				listFiles.PushBackList(&reader)
			}
		}

	} else {
		logger.Warnf("err: %v", err)
	}
	return listFiles
}

func FileExist(filePath string) (exist bool) {
	_, err := os.Stat(filePath)
	return err == nil
}

func FileNotExist(filePath string) (exist bool) {
	return !FileExist(filePath)
}

// SaveDataScriptToFile Файл создается для временного хранения
func SaveDataScriptToFile(logger *zap.SugaredLogger, rqData string, title string) *os.File {
	tFile, err := GetNewFile(logger, fmt.Sprint("LtScriptFile_", title))
	if err != nil {
		logger.Warn("GetNewFile err: ", err)
	}
	logger.Info("-----------tFile Name: ", tFile.Name())

	write, err := tFile.WriteString(rqData)
	if err != nil {
		logger.Warn("WriteString err: ", err)
	}
	logger.Infof("------------tFile dataLen:'%v'; data:'%.100v...'", write, rqData)
	if fileData, er := ReadBytesFromFile(logger, tFile.Name()); er == nil {
		logger.Debugf("ReadAll data from the file: \n'%.100v...'", string(fileData))
	} else {
		logger.Warn(er)
	}
	return tFile
}

func SaveDataAmmoToFile(logger *zap.SugaredLogger, rqData string, fileName string) *os.File {
	tFile, err := GetNewFile(logger, fileName)
	if err != nil {
		logger.Warn("GetNewFile err: ", err)
	}
	logger.Info("-----------tFile Name: ", tFile.Name())

	write, err := tFile.WriteString(rqData)
	if err != nil {
		logger.Warn("WriteString err: ", err)
	}
	logger.Infof("------------tFile dataLen:'%v'; data:'%.100v...'", write, rqData)
	if fileData, er := ReadBytesFromFile(logger, tFile.Name()); er == nil {
		logger.Debugf("ReadAll data from the file: \n'%.100v...'", string(fileData))
	} else {
		logger.Warn(er)
	}

	return tFile
}

func GetNewFile(logger *zap.SugaredLogger, pattern string) (*os.File, error) {
	tFile, err := os.CreateTemp("", fmt.Sprint(pattern, "_"))
	if err != nil {
		logger.Warn("TempFile error: ", err)
	}

	err = tFile.Chmod(0666)
	if err != nil {
		logger.Warn("tFile.Chmod: ", err)
		fi, er := tFile.Stat()
		if er != nil {
			logger.Warn(er)
		}
		fmt.Printf("File permissions  %v %v\n", tFile.Name(), fi.Mode())
	}
	return tFile, err
}
