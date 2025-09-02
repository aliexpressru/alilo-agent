package model

const (
	K6HttpReqs        = "http_reqs"
	K6HttpReqsFailed  = "http_req_failed"
	K6HttpReqDuration = "http_req_duration"
	K6Vus             = "vus_max"
	K6DataSent        = "data_sent"
	K6DataReceived    = "data_data_received"
)

type MetricsResponse struct {
	Data Data `json:"data"`
}

type Data []struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Attributes struct {
		Type     string      `json:"type"`
		Contains string      `json:"contains"`
		Tainted  interface{} `json:"tainted"`
		Sample   struct {
			Count int     `json:"count"`
			Rate  float64 `json:"rate"`
			Value int     `json:"value"`
			Avg   float64 `json:"avg"`
			Max   float64 `json:"max"`
			Med   float64 `json:"med"`
			Min   float64 `json:"min"`
			P90   float64 `json:"p(90)"`
			P95   float64 `json:"p(95)"`
		} `json:"sample"`
	} `json:"attributes"`
}

type Metrics struct {
	Rps      string `json:"rps,omitempty"`
	Rt90P    string `json:"rt90p,omitempty"`
	Rt95P    string `json:"rt95p,omitempty"`
	RtMax    string `json:"rtMax,omitempty"`
	Rt99P    string `json:"rt99p,omitempty"`
	Failed   string `json:"failed,omitempty"`
	Vus      string `json:"vus,omitempty"`
	Sent     string `json:"sent,omitempty"`
	Received string `json:"received,omitempty"`
}
