package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/niwla23/ohfuck/config"
	mqtt_module "github.com/niwla23/ohfuck/modules/mqtt"
	scriptrunner_module "github.com/niwla23/ohfuck/modules/scriptrunner"

	"github.com/niwla23/ohfuck/storage"
	"github.com/niwla23/ohfuck/types"
	"github.com/prometheus/alertmanager/template"
	"golang.org/x/exp/slices"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//go:embed fake_frontend/*
var frontendFs embed.FS

func main() {
	log.Println("[main] starting up..")
	monitorNames := []string{}
	for _, monitor := range config.AppConfig.Monitors {
		monitorNames = append(monitorNames, monitor.Name)
	}

	log.Printf("[main] %d monitors loaded.", len(monitorNames))

	go mqtt_module.StartMQTTHandler()
	go scriptrunner_module.StartScriptRunnerModule()

	app := fiber.New(fiber.Config{AppName: "OhFuck", DisableStartupMessage: true})
	log.Println("[http] loaded app config")
	app.Use(cors.New())
	log.Println("[http] loaded CORS middleware")

	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(frontendFs),
		Browse:     false,
		PathPrefix: "fake_frontend",
	}))
	log.Println("[http] loaded embedded frontend route")

	handleReport := func(c *fiber.Ctx) error {
		monitorName := c.Params("monitorName")
		state := string(c.Params("state"))
		state = strings.TrimSpace(state)
		state = strings.ToUpper(state)
		reason := ""

		if t := strings.TrimSpace(string(c.Body())); t != "" {
			reason = t
		}
		if t := c.GetReqHeaders()["Reason"]; t != "" {
			reason = t
		}
		if t := c.Query("reason"); t != "" {
			reason = t
		}

		if state != "UP" && state != "DOWN" {
			return c.Status(400).SendString("Invalid state")
		}
		if !slices.Contains(monitorNames, monitorName) {
			return c.Status(400).SendString("Invalid monitor name")
		}

		log.Printf("[http] received report for %s, STATE: %s", monitorName, state)
		storage.StoreMonitorState(monitorName, types.MonitorState{Up: state == "UP", Reason: reason, LastReportTime: time.Now()})
		return c.Status(200).SendString("OK")
	}

	app.Post("/api/report/:monitorName/:state", func(c *fiber.Ctx) error {
		return handleReport(c)
	})

	app.Get("/api/report/:monitorName/:state", func(c *fiber.Ctx) error {
		return handleReport(c)
	})

	app.Post("/api/alertmanager", func(c *fiber.Ctx) error {
		// parse request body
		data := new(template.Data)
		if err := c.BodyParser(data); err != nil {
			return err
		}

		log.Println("[http:alertmanager] received alertmanager alerts")
		for _, alert := range data.Alerts {
			monitorName := alert.Labels["ohfuck_name"]
			if monitorName == "" {
				continue
			}
			storage.StoreMonitorState(monitorName, types.MonitorState{Up: alert.Status != "firing", Reason: alert.Annotations["description"], LastReportTime: time.Now()})
		}
		log.Println("[http:alertmanager] end alertmanager alerts")

		return c.SendString("ok")
	})

	app.Get("/api/monitors", func(c *fiber.Ctx) error {
		states := []types.MonitorState{}
		for _, monitorConfig := range config.AppConfig.Monitors {
			monitorState, err := storage.GetMonitorState(monitorConfig.Name)
			check(err)

			timeSinceLastReport := time.Since(monitorState.LastReportTime)
			timeout := monitorConfig.ReportTimeout
			timeoutHit := timeSinceLastReport > timeout

			if monitorState.Up && timeoutHit && monitorConfig.ReportTimeout != 0 {
				monitorState.Up = false
				monitorState.Reason = fmt.Sprintf("No report received for %s", timeout)
			}

			states = append(states, monitorState)
		}

		return c.Status(200).JSON(states)
	})

	log.Println("[http] now listening on port 3000")
	log.Fatal(app.Listen(":3000"))
}
