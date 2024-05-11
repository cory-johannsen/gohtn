package config

type Config struct {
	AssetRoot     string `json:"assetRoot"`
	ConditionPath string `json:"conditionPath"`
	SensorPath    string `json:"sensorPath"`
	TaskPath      string `json:"taskPath"`
	TaskGraphPath string `json:"taskGraphPath"`
	MethodPath    string `json:"methodPath"`
}
