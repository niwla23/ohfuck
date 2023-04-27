package mqtt

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

func subscribeTopicsFromConfig(client MQTT.Client) {
	for _, monitor := range config.AppConfig.Monitors {
		if monitor.MQTT.Topic == "" {
			continue
		}
		if token := client.Subscribe(monitor.MQTT.Topic, byte(0), nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
	}
}

func StartMQTTHandler() {
	opts := MQTT.NewClientOptions().
		AddBroker(config.AppConfig.MQTT.Host).
		SetAutoReconnect(true).
		SetKeepAlive(30 * time.Second).
		SetUsername(config.AppConfig.MQTT.User).
		SetPassword(config.AppConfig.MQTT.Password).
		SetClientID("ohfuck" + randToken(8))

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		message := string(msg.Payload())
		log.Printf("[mqtt] received message, TOPIC: %s MESSAGE: %s\n", msg.Topic(), message)

		for _, monitor := range config.AppConfig.Monitors {
			if monitor.MQTT.Topic == msg.Topic() {
				monitorState := types.MonitorState{Up: true, Reason: "MQTT", LastReportTime: time.Now()}
				if message == monitor.MQTT.UpMessage || monitor.MQTT.UpMessage == "" {
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

	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		log.Printf("[mqtt] connection lost! error: %v\n", err)
	})

	opts.SetReconnectingHandler(func(c MQTT.Client, co *MQTT.ClientOptions) {
		log.Printf("[mqtt] attempting to reconnect\n")
	})

	opts.SetOnConnectHandler(func(c MQTT.Client) {
		log.Printf("[mqtt] sucessfully connected!")
		subscribeTopicsFromConfig(c)
		log.Println("[mqtt] subscribed all configured topics")
	})

	// create a MQTT client, panic on fail
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	log.Printf("[mqtt] created client for %s\n", config.AppConfig.MQTT.Host)

	// keep goroutine alive
	for {
		time.Sleep(1 * time.Second)
	}
}
