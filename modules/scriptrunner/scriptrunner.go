package scriptrunner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/niwla23/ohfuck/config"
	"github.com/niwla23/ohfuck/storage"
	"github.com/niwla23/ohfuck/types"
)

type ScriptResult struct {
	Up     bool   `json:"up"`
	Reason string `json:"reason"`
}

func runScript(name string, args []string) (ScriptResult, error) {
	cmd := exec.Command(config.AppConfig.ScriptRunner.Path+"/"+name, args...)

	// capture the output
	var out bytes.Buffer
	cmd.Stdout = &out

	// wait for the script to finish
	err := cmd.Run()
	if err != nil {
		return ScriptResult{}, err
	}
	lines := strings.Split(out.String(), "\n")

	// find last non-empty line
	resultString := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] != "" {
			resultString = lines[i]
			break
		}
	}

	var scriptResult ScriptResult
	if err := json.Unmarshal([]byte(resultString), &scriptResult); err != nil {
		return scriptResult, err
	}

	return scriptResult, nil
}

func setLastRuntime(monitorName string, timeX time.Time) error {
	return storage.SetValue(fmt.Sprintf("scriptrunner.%s.lastRuntime", monitorName), strconv.FormatInt(timeX.UnixMilli(), 10), 0)
}

func getLastRuntime(monitorName string) (time.Time, error) {
	x, err := storage.GetValue(fmt.Sprintf("scriptrunner.%s.lastRuntime", monitorName))
	if err != nil {
		return time.Time{}, err
	}
	lastRuntime, err := strconv.ParseInt(x, 10, 64)
	return time.UnixMilli(lastRuntime), err
}

func StartScriptRunnerModule() {
	log.Println("[scriptrunner] started")
	for {
		for _, monitor := range config.AppConfig.Monitors {
			if monitor.ScriptRunner.Script == "" {
				continue
			}
			lastRuntime, _ := getLastRuntime(monitor.Name)

			if lastRuntime.Add(monitor.ScriptRunner.Interval).Before(time.Now()) {
				log.Printf("[scriptrunner] running script '%s' for monitor %s\n", monitor.ScriptRunner.Script, monitor.Name)
				scriptResult, err := runScript(monitor.ScriptRunner.Script, monitor.ScriptRunner.Args)
				if err != nil {
					log.Printf("[scriptrunner] error running script: %v: ", err)
				}
				storage.StoreMonitorState(monitor.Name, types.MonitorState{Up: scriptResult.Up, Reason: scriptResult.Reason})

				err = setLastRuntime(monitor.Name, time.Now())
				if err != nil {
					log.Printf("[scriptrunner] error getting last runtime: %v: ", err)
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}
