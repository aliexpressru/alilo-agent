package http_utils

import (
	"bytes"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// GetAndConvertBodyToString Читаем тело из POST запроса
func GetAndConvertBodyToString(logger *zap.SugaredLogger, r *http.Request) string {
	return GetAndConvertReadToString(logger, r.Body)
}

// GetAndConvertReadToString Получаем строку из ио.ридера
func GetAndConvertReadToString(logger *zap.SugaredLogger, reader io.Reader) string {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		logger.Warn("Ошибка чтения строки из io.Reader")
	}
	rqData := buf.String()
	logger.Debugf("buf.ReadFrom: '%+v'", rqData)
	return rqData
}

// GetAndConvertReadToBytes Получаем байты из ио.ридера
func GetAndConvertReadToBytes(logger *zap.SugaredLogger, reader io.Reader) []byte {
	logger.Infof("Обработка тела сообщения")
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		logger.Warn("Ошибка чтения строки из io.Reader")
	}
	logger.Debugf("buf.ReadFrom: '%+v'", buf.String())
	return buf.Bytes()
}

func Foreword(logger *zap.SugaredLogger, r *http.Request) (rqData string) {
	rqData = GetAndConvertBodyToString(logger, r)
	logger.Debugf("Foreword Rq: \n RequestURI: '%+v'\n Header: '%+v'\n RequestBody: '%v'\n",
		r.RequestURI, r.Header, rqData)
	return rqData
}

func Favicon(w http.ResponseWriter, r *http.Request) {
	Foreword(zap.S(), r)
	w.Header().Add("Cache-Control", "public, max-age=31536000")
	http.ServeFile(w, r, "resources/Favicon.ico")
}

func Get(logger *zap.SugaredLogger, uri string) (body []byte, err error) {
	logger.Infof("----Start saved script 3")
	defer func() {
		if r := recover(); r != nil {
			logger.Warnf("Recovered in Get: %v", r)
		}
	}()
	logger.Infof("----Start saved script 3")
	logger.Infof("---------- START http.Get: '%v'", uri)
	resp, er := http.Get(uri)
	defer func(resp *http.Response) {
		err = resp.Body.Close()
		if err != nil {
			logger.Warnf("Body.Close: %v", err)
		}
	}(resp)
	if er != nil {
		logger.Warnf("http.Get Error: %+v", er)
		return body, er
	}

	body = GetAndConvertReadToBytes(logger, resp.Body)

	logger.Debugf("---------- http.Get: '%.50v'", string(body))
	return body, er
}

func GetWithHeaders(logger *zap.SugaredLogger, uri string, headers map[string]string) string {
	logger.Infof("--Start saved script 3")
	defer func() {
		if r := recover(); r != nil {
			logger.Warnf("Recovered in GetWithHeaders(): %v", r)
		}
	}()
	logger.Infof("------- START http.Get: '%v'", uri)
	client := http.DefaultClient
	req, err := http.NewRequest("GET", uri, nil)
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	resp, err := client.Do(req)
	defer func(resp *http.Response) {
		err = resp.Body.Close()
		if err != nil {
			logger.Warn("http.GetWithHeaders Error: ", err)
		}
	}(resp)
	if err != nil {
		logger.Warn("http.GetWithHeaders Error: ", err)
	}
	body := GetAndConvertReadToString(logger, resp.Body)

	logger.Debugf("---------- http.Get: '%.50v'", body)
	return body
}
