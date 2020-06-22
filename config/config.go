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
		cfg.ServName = "filesync"
		return err
	}
	cfg.Host, _  		= cfgObj.String("filesync", "host")
	cfg.Port, _     	= cfgObj.Int("filesync", "port")
	cfg.User, _     	= cfgObj.String("filesync", "user")
	cfg.Passwd, _   	= cfgObj.String("filesync", "passwd")
	cfg.Interval, _ 	= cfgObj.Int("filesync", "interval")
	cfg.Retry, _ 		= cfgObj.Int("filesync", "retry")
	cfg.Path, _     	= cfgObj.String("filesync", "path")
	cfg.RemotePath, _ 	= cfgObj.String("filesync", "remote_path")
	cfg.LogLevel, _ 	= cfgObj.Int("filesync", "log_level")
	cfg.LogSize, _  	= cfgObj.Int("filesync", "log_size")
	cfg.LogCount, _ 	= cfgObj.Int("filesync", "log_count")
	cfg.Reverse, _ 		= cfgObj.Bool("filesync", "new_file_first")
	cfg.Timeout, _ 		= cfgObj.Int("filesync", "timeout")
	cfg.ServName, _		= cfgObj.String("filesync", "serv_name")
	duration, _			:= cfgObj.String("filesync", "work_during")
	cfg.Protocol, _		= cfgObj.String("filesync", "protocol")
	cfg.TransMode, _ 	= cfgObj.String("filesync", "trans_mode")
	cfg.Buffer, _ 		= cfgObj.Int("filesync", "buffer")

	cfg.From = "00:00"
	cfg.To = "23:59"
	hours := strings.Split(duration, "-")
	if len(hours) > 1 {
		cfg.From = hours[0]
		cfg.To = hours[1]
	}
	return nil
}
