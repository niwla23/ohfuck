package consumers

import "github.com/niwla23/ohfuck/consumers/ntfy"

func GetAllConsumers() []func() {
	var consumerFunctions []func()
	consumerFunctions = append(consumerFunctions, ntfy.ProcessEvent)
	return consumerFunctions
}

func NotifyAllConsumers() {
	consumerFunctions := GetAllConsumers()
	for _, c := range consumerFunctions {
		c()
	}
}
