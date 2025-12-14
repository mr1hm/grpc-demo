package config

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	*logrus.Logger
	UserServicePort    string
	GatewayServicePort string
}

func New(userServicePort, gatewayServicePort string) *Config {
	return &Config{
		Logger:             logrus.New(),
		UserServicePort:    userServicePort,
		GatewayServicePort: gatewayServicePort,
	}
}
