package biz

import "chronoFlow-admin/internal/conf"

func NewJobRunConfig(server *conf.Server, security *conf.Security) JobRunConfig {
	config := JobRunConfig{}
	if server != nil {
		config.PublicBaseURL = server.PublicBaseUrl
	}
	if security != nil {
		config.CallbackToken = security.CallbackToken
	}
	return config
}

func NewCallbackConfig(logs *conf.Logs) CallbackConfig {
	config := CallbackConfig{}
	if logs != nil {
		config.MaxLogBytes = logs.MaxLogBytes
	}
	return config
}
