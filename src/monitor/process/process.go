package process

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

var lastRestart = make(map[string]time.Time)

func MonitorProcess(logger zerolog.Logger, processStop chan struct{}) {
	logger.Info().Msg("Process Monitor thread started")

	HaltTime := os.Getenv("PROCESS_HALT_TIME")
	if HaltTime == "" {
		HaltTime = "10"
	}
	haltDuration, err := strconv.Atoi(HaltTime)
	if err != nil {
		logger.Error().Msgf("Error converting PROCESS_HALT_TIME to an integer:", err)
		return
	}
	ticker := time.NewTicker(time.Second * time.Duration(haltDuration))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			monitorProcesses(logger)

		case <-processStop:
			logger.Info().Msg("Process thread exiting")
			return
		}
	}
}

func monitorProcesses(logger zerolog.Logger) {
	// dummy process to monitor
	desiredProcess := DbPm2Process{
		Name:    "test",
		Command: "cargo run",
		PWD:     "/Users/root/code/logging",
		Env:     []map[string]string{{"RUST_ENV": "production"}},
	}

	// Todo: get the list of process to monitor from the db
	desiredProcesses := []DbPm2Process{}
	desiredProcesses = append(desiredProcesses, desiredProcess)

	processesStatus := IsProcessesRunning(desiredProcesses, logger)

	for _, p := range processesStatus {

		if p.Status == "stopped" {
			if time.Since(lastRestart[p.Name]) < 2*time.Second {
				continue
			}

			logger.Warn().Msgf("Restarting %s (status: %s, pid: %d)", p.Name, p.Status, p.PID)
			RestartProcess(p.PmId, p.Name, logger)
			lastRestart[p.Name] = time.Now()

			// Allow time for PM2 to update status
			time.Sleep(2 * time.Second)

		} else if p.Status == "start" {
			ProcessStart := os.Getenv("PROCESS_START")
			if ProcessStart == "FALSE" {
				logger.Info().Msg("PROCESS_START is set to false, not starting container")
				continue
			}

			logger.Info().Msgf("Starting %s", p.Name)
			StartProcess(desiredProcess, logger)
		}
	}

}

func IsProcessesRunning(desiredProcess []DbPm2Process, logger zerolog.Logger) []ProcessStatus {
	processes := GetPm2Processes(logger)

	processStatus := []ProcessStatus{}

	for _, desired := range desiredProcess {
		found := bool(false)

		for _, p := range processes {
			if p.Name == desired.Name {
				if p.Status == "online" || p.Status == "launching" {
					processStatus = append(processStatus, ProcessStatus{
						Status:  "online",
						PID:     p.PID,
						Name:    p.Name,
						PmId:    p.PmID,
						Command: p.Command,
						Env:     p.Env,
						PWD:     p.PWD,
					})
				} else {
					logger.Error().Msgf("Process %s is not running, Status: %s", p.Name, p.Status)
					processStatus = append(processStatus, ProcessStatus{
						Status:  "stopped",
						PID:     p.PID,
						Name:    p.Name,
						PmId:    p.PmID,
						Command: p.Command,
						Env:     p.Env,
						PWD:     p.PWD,
					})
				}
				found = true
				break
			}
		}

		if !found {
			logger.Error().Msgf("Process %s is not found in Pm2", desired.Name)
			processStatus = append(processStatus, ProcessStatus{
				Status:  "start",
				PID:     0,
				Name:    desired.Name,
				PmId:    0,
				Command: desired.Command,
				Env:     desired.Env,
				PWD:     desired.PWD,
			})
		}
	}
	return processStatus
}

func GetPm2Processes(logger zerolog.Logger) []Pm2Process {
	var processes []Pm2Process
	cmd := exec.Command("pm2", "jlist")

	output, err := cmd.Output()
	if err != nil {
		logger.Error().Msgf("Error getting pm2 processes: %s", err)
		return nil
	}

	rawProcesses := []RawPm2Process{}
	if err := json.Unmarshal(output, &rawProcesses); err != nil {
		logger.Error().Msgf("Error unmarshalling pm2 processes: %s", err)
		return nil
	}

	for _, p := range rawProcesses {
		command := p.Pm2Env.Args[len(p.Pm2Env.Args)-1]

		processes = append(processes, Pm2Process{
			PmID:    p.PmID,
			Name:    p.Name,
			PID:     p.PID,
			Status:  p.Pm2Env.Status,
			Command: command,
			PWD:     p.Pm2Env.PWD,
			Env:     p.Pm2Env.Env,
		})
	}

	return processes
}

func RestartProcess(pmID int, processName string, logger zerolog.Logger) {
	cmd := exec.Command("pm2", "restart", fmt.Sprintf("%d", pmID))
	if err := cmd.Run(); err != nil {
		logger.Error().Msgf("Error restarting process: %s", err)
	}
}

func StartProcess(process DbPm2Process, logger zerolog.Logger) {
	cmd := exec.Command("pm2", "start", process.Command, "--name", process.Name)
	cmd.Dir = process.PWD

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error().
			Err(err).
			Str("output", string(output)).
			Str("directory", process.PWD).
			Msg("Failed to start process")
		return
	}

	logger.Info().
		Str("name", process.Name).
		Str("directory", process.PWD).
		Msg("Process started successfully")

}
