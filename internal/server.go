package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aliexpressru/alilo-agent/internal/model"
	"github.com/aliexpressru/alilo-agent/pkg/utils/http_utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

var (
	serverPort string
)

func middleware(label string, next http.Handler) http.Handler {
	logger.Infof("Registr HandleFunc: %s ", label)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("_-_-_-_-_ Handler %s ", label)
		defer logger.Infof("middleware completed: %s ", label)
		defer func() {
			if err := recover(); err != nil {
				logger.Warnf("_Middleware recovered in '%v': '%v'", label, err)
				var returnErr error
				switch x := err.(type) {
				case string:
					returnErr = errors.New(x)
				case error:
				default:
					returnErr = fmt.Errorf("unknown panic: %v", err)
				}
				resp := &model.Response{}
				resp.SetError(returnErr.Error())
				returnErrorResponse(w, resp, returnErr)
				return
			}
		}()
		w.Header().Set("Cache-Control", "no-cache, no-store")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("accept", "*/*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		next.ServeHTTP(w, r)
	})
}

func addCors() (http.Handler, *http.ServeMux) {
	mux := http.NewServeMux()
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		//AllowedOrigins:     []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "DELETE", "POST", "PUT"},
		//AllowedHeaders:     []string{"*"},
		//OptionsPassthrough: true,
		//Debug:              true,
	})
	handler := c.Handler(mux)
	return handler, mux
}

func PrepareServerAPI(mux *http.ServeMux) {
	//mux.HandleFunc("/", rootPage)
	mux.Handle("/metrics", promhttp.Handler())
	favicon := http.HandlerFunc(http_utils.Favicon)
	mux.Handle("/api/v1/Favicon.ico", middleware("FaviconHandler", favicon))
	handlerSaveScript := http.HandlerFunc(SaveScript)
	mux.Handle("/api/v1/saveScript", middleware("SaveScriptHandler", handlerSaveScript))

	GetMetricsAgentUtilization := http.HandlerFunc(GetMetricsAgentResourceUtilization)
	mux.Handle("/api/v1/agent/metrics", middleware("GetMetricsAgentResourceUtilization", GetMetricsAgentUtilization))

	handlerStartCommand := http.HandlerFunc(StartCommand)
	mux.Handle("/api/v1/start", middleware("StartCommandHandler", handlerStartCommand))
	handlerStopCommand := http.HandlerFunc(StopCommand)
	mux.Handle("/api/v1/stop", middleware("StopCommandHandler", handlerStopCommand))

	HandlerGetAllTasks := http.HandlerFunc(GetAllTasks)
	mux.Handle("/api/v1/getAllTasks", middleware("GetAllTasksHandler", HandlerGetAllTasks))
	handlerGetTask := http.HandlerFunc(GetTask)
	mux.Handle("/api/v1/getTask", middleware("GetTaskHandler", handlerGetTask))
	handlerGetTaskLogs := http.HandlerFunc(GetTaskLogs)
	mux.Handle("/api/v1/getTaskLogs", middleware("GetTaskLogsHandler", handlerGetTaskLogs))
	handlerGetStatus := http.HandlerFunc(GetStatus)
	mux.Handle("/api/v1/getStatus", middleware("GetStatusHandler", handlerGetStatus))
}

func returnErrorResponse(w http.ResponseWriter, respErrorResponse model.IResponse, err error) {
	logger.Warnf("ReturnErrorResponse: '%v'", err.Error())
	respErrorResponse.SetError(err.Error())
	logger.Warnf("ReturnErrorResponse: '%v'", respErrorResponse)
	returnResponse(w, respErrorResponse, ResponseStatusError)
}

func returnResponse(w http.ResponseWriter, resp model.IResponse, status string) {
	logger.Infof("Return status response: '%v'", status)
	resp.SetStatus(status)
	logger.Debugf("Return response: '%+v'", resp)
	indentResp, er := json.MarshalIndent(&resp, "", "	")
	if er != nil {
		logger.Warn("Marshal resp error: ", er)
	}
	logger.Debugf("Marshal resp: %s", indentResp)
	_, err := fmt.Fprintf(w, "%s", indentResp)
	if err != nil {
		logger.Warn(err)
	}
}

func runEngineAgent(handler http.Handler) {
	if cfg.Tls.Use {
		fmt.Printf("Server HTTPS launching...")
		if err := http.ListenAndServeTLS(fmt.Sprintf("0.0.0.0:%v", cfg.ServerPort), cfg.Tls.CertFile, cfg.Tls.KeyFile, handler); err != nil {
			fmt.Println("Socket is busy\n", err.Error())
			logger.Panic(err.Error())
		} else {
			logger.Info("Agent with TLS started")
		}
	} else {
		fmt.Println("Server HTTP launching...")
		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", cfg.ServerPort), handler); err != nil {
			fmt.Println("Socket is busy\n", err.Error())
			logger.Panic(err.Error())
		} else {
			logger.Info("Agent without TLS started")
		}
	}
}
