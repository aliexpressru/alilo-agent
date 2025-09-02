package model

// Config Конфиг всего эмулятора
type Config struct {
	ServerPort string `json:"serverPort"`
	LogLevel   string `json:"logLevel"`
	Tls        TLS    `json:"tls"` // Параметр шифрации сервера
}

type TLS struct {
	Use      bool   `json:"use"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}
