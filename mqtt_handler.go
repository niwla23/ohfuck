package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/niwla23/ohfuck/config"
	"github.com/niwla23/ohfuck/storage"
	"github.com/niwla23/ohfuck/types"
)

// randToken generates a random hex value.
func randToken(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		panic("so for some reason we cant get random data. Your system is probably fucked up quite badly.")
	}
	return hex.EncodeToString(bytes)
}

func startMQTTHandler() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.AppConfig.MQTT.Host)
	opts.SetUsername(config.AppConfig.MQTT.User)
	opts.SetPassword(config.AppConfig.MQTT.Password)
	opts.SetClientID("ohfuck" + randToken(8))

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		message := string(msg.Payload())
		log.Printf("received mqtt, TOPIC: %s MESSAGE: %s\n", msg.Topic(), message)

		for _, monitor := range config.AppConfig.Monitors {
			if monitor.MQTT.Topic == msg.Topic() {
				monitorState := types.MonitorState{Up: true, Reason: "MQTT", LastReportTime: time.Now()}
				if message == monitor.MQTT.UpMessage {
					monitorState.Up = true
				} else if message == monitor.MQTT.DownMessage {
					monitorState.Up = false
				} else {
					continue
				}
				storage.StoreMonitorState(monitor.Name, monitorState)
			}
		}
	})

	// create a MQTT client, panic on fail
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// subscribe all configured topics
	for _, monitor := range config.AppConfig.Monitors {
		if monitor.MQTT.Topic == "" {
			continue
		}
		if token := client.Subscribe(monitor.MQTT.Topic, byte(0), nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
	}

	// keep goroutine alive
	for {
		time.Sleep(1 * time.Second)
	}
}
