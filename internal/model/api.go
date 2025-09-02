package model

type Request struct {
	Pid int `json:"pid"`
}
type RequestGetTaskLogs struct {
	Pid  int    `json:"pid"`
	Len  int    `json:"len"`
	Name string `json:"name"`
	Head bool   `json:"head"`
}

type IResponse interface {
	SetStatus(status string)
	SetError(status string)
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Pid    int    `json:"pid"`
	Task   *Task  `json:"task"`
}

func (r *Response) SetStatus(status string) {
	r.Status = status
}
func (r *Response) SetError(error string) {
	r.Error = error
}

/*{
	"pid": [
		42890
	],
	"status": "Success",
	"error": ""
}*/

type ResponseGetAllTasks struct {
	Status string        `json:"status"`
	Error  string        `json:"error"`
	Tasks  map[int]*Task `json:"tasks"`
}

func (r *ResponseGetAllTasks) SetStatus(status string) {
	r.Status = status
}
func (r *ResponseGetAllTasks) SetError(error string) {
	r.Error = error
}

type ResponseGetStatus struct {
	Status  string   `json:"status"`
	Error   string   `json:"error"`
	Task    *Task    `json:"task"`
	Metrics *Metrics `json:"metrics"`
}

func (r *ResponseGetStatus) SetStatus(status string) {
	r.Status = status
}
func (r *ResponseGetStatus) SetError(error string) {
	r.Error = error
}

type ResponseSaveScript struct {
	Status     string `json:"status"`
	Error      string `json:"error"`
	PathScript string `json:"pathScript"`
}

func (r *ResponseSaveScript) SetStatus(status string) {
	r.Status = status
}
func (r *ResponseSaveScript) SetError(error string) {
	r.Error = error
}

type ResponseUtilization struct {
	Status                  string                  `json:"status"`
	Error                   string                  `json:"error"`
	PercentAgentUtilization PercentAgentUtilization `json:"agentUtilization"`
}

func (r *ResponseUtilization) SetStatus(status string) {
	r.Status = status
}
func (r *ResponseUtilization) SetError(error string) {
	r.Error = error
}
