package storage

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/niwla23/ohfuck/config"
	"github.com/niwla23/ohfuck/types"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	Addr: config.AppConfig.RedisHost,
})

func init() {
	if err := rdb.Ping(ctx); err.Err() != nil {
		log.Println(err)
		log.Println("Unable to connect to redis. Exiting...")
		os.Exit(1)
	}
}

func StoreMonitorState(monitorName string, monitorState types.MonitorState) error {
	encodedData, err := json.Marshal(monitorState)
	if err != nil {
		return err
	}

	log.Printf("storing new monitor state, NAME: %s, UP: %v", monitorName, monitorState.Up)
	return rdb.Set(ctx, monitorName, encodedData, 0).Err()
}

func GetMonitorState(monitorName string) (types.MonitorState, error) {
	monitorState := types.MonitorState{}

	found := false
	friendlyName := ""
	for _, monitor := range config.AppConfig.Monitors {
		if monitorName == monitor.Name {
			found = true
			friendlyName = monitor.FriendlyName
		}
	}

	if !found {
		return monitorState, errors.New("monitor not found")
	}
	monitorState.Name = monitorName
	monitorState.FriendlyName = friendlyName

	raw, err := rdb.Get(ctx, monitorName).Result()
	if err != nil {
		return monitorState, nil
	}

	err = json.Unmarshal([]byte(raw), &monitorState)
	if err != nil {
		return monitorState, err
	}

	monitorState.Name = monitorName
	monitorState.FriendlyName = friendlyName

	return monitorState, nil
}
