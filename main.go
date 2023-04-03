package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/storage/sqlite3"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	Monitors []struct {
		Name          string        `yaml:"name"`
		FriendlyName  string        `yaml:"friendlyName"`
		ReportTimeout time.Duration `yaml:"reportTimeout"`
	} `yaml:"monitors"`
}

type MonitorState struct {
	Name           string    `json:"name"`
	FriendlyName   string    `json:"friendlyName"`
	Up             bool      `json:"up"`
	Reason         string    `json:"reason"`
	LastReportTime time.Time `json:"lastReportTime"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//go:embed fake_frontend/*
var frontendFs embed.FS

func main() {
	store := sqlite3.New(sqlite3.Config{
		Database:        "./ohfuck.sqlite3",
		Table:           "fiber_storage",
		Reset:           false,
		GCInterval:      10 * time.Second,
		MaxOpenConns:    100,
		MaxIdleConns:    100,
		ConnMaxLifetime: 1 * time.Second,
	})

	var appConfig AppConfig

	fp := os.Getenv("OHFUCK_CONFIG_FILE")
	if fp == "" {
		panic("OHFUCK_CONFIG_FILE is not set")
	}

	dat, err := os.ReadFile(fp)
	check(err)
	err = yaml.Unmarshal(dat, &appConfig)
	check(err)
	fmt.Println(appConfig.Monitors)
	monitorNames := []string{}
	for _, monitor := range appConfig.Monitors {
		monitorNames = append(monitorNames, monitor.Name)
	}
	app := fiber.New(fiber.Config{AppName: "OhFuck"})

	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(frontendFs),
		Browse:     true,
		PathPrefix: "fake_frontend",
	}))

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
		encodedData, err := json.Marshal(MonitorState{Up: state == "UP", Reason: reason, LastReportTime: time.Now()})
		check(err)
		store.Set(monitorName, encodedData, 0)
		return c.SendString("OK")
	}

	app.Post("/api/report/:monitorName/:state", func(c *fiber.Ctx) error {
		return handleReport(c)
	})

	app.Get("/api/report/:monitorName/:state", func(c *fiber.Ctx) error {
		return handleReport(c)
	})

	app.Get("/api/monitors", func(c *fiber.Ctx) error {
		states := []MonitorState{}
		for _, monitorConfig := range appConfig.Monitors {
			raw, err := store.Get(monitorConfig.Name)

			if len(raw) == 0 {
				states = append(states, MonitorState{Name: monitorConfig.Name, FriendlyName: monitorConfig.FriendlyName, Up: false, Reason: "No Information"})
				continue
			}

			check(err)
			var monitorState MonitorState
			err = json.Unmarshal(raw, &monitorState)
			check(err)
			monitorState.Name = monitorConfig.Name
			monitorState.FriendlyName = monitorConfig.FriendlyName

			timeSinceLastReport := time.Since(monitorState.LastReportTime)
			timeout := monitorConfig.ReportTimeout
			timeoutHit := timeSinceLastReport > timeout

			if monitorState.Up && timeoutHit && monitorConfig.ReportTimeout != 0 {
				monitorState.Up = false
				monitorState.Reason = fmt.Sprintf("No report received for %s", timeout)
			}

			states = append(states, monitorState)
		}

		return c.JSON(states)
	})

	log.Fatal(app.Listen(":3000"))
}
