package config

import "github.com/SALT-Indonesia/salt-pkg/logmanager"

type ServerConfig struct {
	Port    string
	AppName string
}

func NewDefaultLogManager(appName string, opts ...logmanager.Option) *logmanager.Application {
	defaultOpts := []logmanager.Option{
		logmanager.WithAppName(appName),
	}
	defaultOpts = append(defaultOpts, opts...)
	return logmanager.NewApplication(defaultOpts...)
}