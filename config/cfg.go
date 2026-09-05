package config

import (
	"fmt"
	"io"

	"go.yaml.in/yaml/v4"
)

type Service struct {
	Name     string   `yaml:"name"`
	Replicas []string `yaml:"replicas"`
	Strategy string   `yaml:"strategy"`
	Weights  []uint32 `yaml:"weights"`
}
type Config struct {
	Services []*Service `yaml:"services"`
	//Todo: remember to deal with this defualt
	DefaultService string
}

func LoadConfig(reader io.Reader) (*Config, error) {
	var config Config
	err := yaml.NewDecoder(reader).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("loadconfig: %w", err)
	}
	return &config, nil
}
