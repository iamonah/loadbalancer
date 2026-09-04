package config

type Service struct {
	Name     string
	Replicas []string
	Strategy string
}

type Config struct {
	Services       map[string]*Service
	DefaultService string
}

func LoadConfig() *Config {
	return &Config{
		Services: map[string]*Service{
			"TestService": {
				Name:     "TestService",
				Replicas: []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"},
				Strategy: "round-robin",
			},
		},
		DefaultService: "TestService",
	}

}
