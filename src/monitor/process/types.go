package process

type RawPm2Process struct {
	PmID   int    `json:"pm_id"`
	PID    int    `json:"pid"`
	Name   string `json:"name"`
	Pm2Env struct {
		Status     string              `json:"status"`
		PmUptime   int64               `json:"pm_uptime"`
		PmExecPath string              `json:"pm_exec_path"`
		PWD        string              `json:"PWD"`
		Args       []string            `json:"args"`
		Env        []map[string]string `json:"filter_env"`
	} `json:"pm2_env"`
	Monit struct {
		Memory int64   `json:"memory"`
		CPU    float64 `json:"cpu"`
	} `json:"monit"`
}

type Pm2Process struct {
	PmID    int
	Name    string
	PID     int
	Status  string
	PWD     string
	Command string
	Env     []map[string]string
}

type DbPm2Process struct {
	Name    string              `json:"name"`
	Command string              `json:"command"`
	Env     []map[string]string `json:"env"`
	PWD     string              `json:"pwd"`
}

type ProcessStatus struct {
	Status  string
	PID     int
	Name    string
	PmId    int
	Command string
	Env     []map[string]string
	PWD     string
}
