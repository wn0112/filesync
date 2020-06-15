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
		Name:        Cfg.ServName,             //服务显示名称
		DisplayName: Cfg.ServName,             //服务名称
		Description: "Upload or download file to/from ftp/sftp server", //服务描述
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
			err = conn.Close()
			if err != nil {
				Logger.E(err.Error())
			}

			Logger.I("Task completed. %d file(s) were transfered. Sleeping now: %d minute(s)",
						TotalFile, Cfg.Interval)
			time.Sleep(time.Duration(Cfg.Interval) * time.Minute)
	}
}


func GetConnector(protocol string) Connector {
	switch protocol {
	case "sftp":
		msftp := sftp.Sftp{nil,Cfg.Host,Cfg.User,Cfg.Passwd,Cfg.Port}
		return &msftp
	default:
		mftp := ftp.Ftp{nil,Cfg.Host,Cfg.User,Cfg.Passwd,Cfg.Port}
		return &mftp
	}
	return nil
}


func PrintConfig(cfg *Config) {
	Logger.I(strings.Repeat("*", SYMBOL))
	Logger.I("* Server:\t\t\t[ %s:%d ]", cfg.Host, cfg.Port)
	Logger.I("* Protocol:\t\t[ %s ]", cfg.Protocol)
	Logger.I("* Username:\t\t[ %s ]", cfg.User)
	Logger.I("* Transfer mode:\t[ %s ]", cfg.TransMode)
	Logger.I("* Buffer size:\t\t[ %d ]", cfg.Buffer)
	Logger.I("* Local path:\t\t[ %s ]", cfg.Path)
	Logger.I("* Remote path:\t\t[ %s ]", cfg.RemotePath)
	Logger.I("* New file first:\t[ %v ]", cfg.Reverse)
	Logger.I("* Service name:\t[ %s ]", cfg.ServName)
	Logger.I("* Working time:\t[ %s - %s ]", cfg.From, cfg.To)
	Logger.I(strings.Repeat("*", SYMBOL))
}


