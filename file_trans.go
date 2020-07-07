package main

import (
	. "ftp_upload/config"
	"ftp_upload/ftp"
	. "ftp_upload/global"
	"ftp_upload/sftp"
	"github.com/kardianos/service"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var err error

type Program struct{}

func main() {
	svcConfig := &service.Config{
		Name:        Cfg.ServName,       //服务显示名称
		DisplayName: Cfg.ServName,       //服务名称
		Description: "ftp/sftp 数据上传/下载", //服务描述
	}

	prg := &Program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		Logger.E(err.Error())
	}

	if err != nil {
		Logger.E(err.Error())
	}

	if len(os.Args) > 1 {
		for _, arg := range os.Args {
			if arg == "install" {
				s.Install()
				Logger.I("服务安装成功")
				return
			}

			if arg == "uninstall" {
				s.Uninstall()
				Logger.I("服务卸载成功")
				return
			}
		}

	}

	err = s.Run()
	if err != nil {
		Logger.Error(err.Error())
	}

}

func (p *Program) Start(s service.Service) error {
	go p.Run()
	return nil
}

func (p *Program) Stop(s service.Service) error {
	Stop = true
	return nil
}

func (p *Program) Run() {
	var conn Connector
	for {
		TotalFile = 0
		Logger.I(strings.Repeat("=", SYMBOL))
		err = Cfg.Load(filepath.Join(RootPath, CONFIG))
		if err != nil {
			Logger.E(err.Error())
			goto END
		}

		PrintConfig(&Cfg)

		// 验证本地目录是否存在
		_, err = os.Stat(Cfg.Path)
		if err != nil {
			Logger.E(err.Error())
			goto END
		}

		// 获取连接
		conn = GetConnector(Cfg.Protocol)
		err = conn.Connect()
		if err != nil {
			Logger.E(err.Error())
			goto END
		}

		// 工作模式, upload or download
		if strings.Contains(strings.ToLower(Cfg.TransMode), "up") {
			err = conn.UploadDirectory(Cfg.Path, Cfg.RemotePath)
			if err != nil {
				Logger.E("%s -> %s", err.Error(), Cfg.RemotePath)
				goto END
			}
		} else {
			err = conn.DownloadDirectory(Cfg.Path, Cfg.RemotePath)
			if err != nil {
				Logger.E("%s -> %s", err.Error(), Cfg.RemotePath)
				goto END
			}
		}

	END:
		if conn != nil {
			err = conn.Close()
			if err != nil {
				Logger.E(err.Error())
			}
		}

		Logger.I("Task completed. %d file(s) were transfered. Sleeping now: %d minute(s)",
			TotalFile, Cfg.Interval)
		time.Sleep(time.Duration(Cfg.Interval) * time.Minute)
	}
}

func GetConnector(protocol string) Connector {
	switch protocol {
	case "sftp":
		msftp := sftp.Sftp{nil, Cfg.Host, Cfg.User, Cfg.Passwd, Cfg.Port}
		return &msftp
	default:
		mftp := ftp.Ftp{nil, Cfg.Host, Cfg.User, Cfg.Passwd, Cfg.Port}
		return &mftp
	}
	return nil
}

func PrintConfig(cfg *Config) {
	Logger.I(strings.Repeat("*", SYMBOL))
	Logger.I("* %-18s [ %s:%d ]", "Server:", cfg.Host, cfg.Port)
	Logger.I("* %-18s [ %s ]", "Protocol:", cfg.Protocol)
	Logger.I("* %-18s [ %s ]", "Transfer mode:", cfg.TransMode)
	Logger.I("* %-18s [ %s ]", "Username:", cfg.User)
	Logger.I("* %-18s [ %d ]", "Buffer size:", cfg.Buffer)
	Logger.I("* %-18s [ %s ]", "Local path:", cfg.Path)
	Logger.I("* %-18s [ %s ]", "Remote path:", cfg.RemotePath)
	Logger.I("* %-18s [ %v ]", "New file first:", cfg.Reverse)
	Logger.I("* %-18s [ %s - %s ]", "Working time:", cfg.From, cfg.To)
	Logger.I(strings.Repeat("*", SYMBOL))
}
