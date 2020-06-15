package config

import (
	"github.com/robfig/config"
	"strings"
)


type Config struct {
	Host     	string
	Port     	int
	User     	string
	Passwd 	 	string
	Path	 	string
	RemotePath	string
	Interval 	int
	Retry		int
	LogLevel 	int
	LogSize  	int
	LogCount 	int
	Reverse  	bool
	Timeout  	int
	ServName 	string
	Protocol 	string
	Buffer		int
	TransMode   string
	From		string
	To			string
}


func (cfg *Config) Load(file string) error {
	cfgObj, err := config.ReadDefault(file)
	if err != nil {
		cfg.Interval = 5
		return err
	}
	cfg.Host, _  		= cfgObj.String("params", "host")
	cfg.Port, _     	= cfgObj.Int("params", "port")
	cfg.User, _     	= cfgObj.String("params", "user")
	cfg.Passwd, _   	= cfgObj.String("params", "passwd")
	cfg.Interval, _ 	= cfgObj.Int("params", "interval")
	cfg.Retry, _ 		= cfgObj.Int("params", "retry")
	cfg.Path, _     	= cfgObj.String("params", "path")
	cfg.RemotePath, _ 	= cfgObj.String("params", "remote_path")
	cfg.LogLevel, _ 	= cfgObj.Int("params", "log_level")
	cfg.LogSize, _  	= cfgObj.Int("params", "log_size")
	cfg.LogCount, _ 	= cfgObj.Int("params", "log_count")
	cfg.Reverse, _ 		= cfgObj.Bool("params", "new_file_first")
	cfg.Timeout, _ 		= cfgObj.Int("params", "timeout")
	cfg.ServName, _		= cfgObj.String("params", "serv_name")
	duration, _			:= cfgObj.String("params", "work_during")
	cfg.Protocol, _		= cfgObj.String("params", "protocol")
	cfg.TransMode, _ 	= cfgObj.String("params", "trans_mode")
	cfg.Buffer, _ 		= cfgObj.Int("params", "buffer")

	cfg.From = "00:00"
	cfg.To = "23:59"
	hours := strings.Split(duration, "-")
	if len(hours) > 1 {
		cfg.From = hours[0]
		cfg.To = hours[1]
	}
	return nil
}

