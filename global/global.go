package global

import (
	. "ftp_upload/config"
	"github.com/aiwuTech/fileLogger"
	"github.com/jlaffaye/ftp"
	"os"
	"path/filepath"
	"time"
)

type FtpEntryList []*ftp.Entry
type FileInfoList []os.FileInfo
type Connector interface {
	Connect() error
	UploadDirectory(string, string) error
	DownloadDirectory(string, string) error
	Close() error
}

const (
	CONFIG     = "configure.ini"
	BUFFERSIZE = 1024
	SYMBOL     = 60
)

var (
	Cfg       Config
	RootPath  string
	Logger    *fileLogger.FileLogger
	TotalFile uint
	Buff      []byte
	Stop      = false
)

func init() {
	self, _ := filepath.Abs(os.Args[0])
	RootPath = filepath.Dir(self)
	err := Cfg.Load(filepath.Join(RootPath, CONFIG))
	Logger = fileLogger.NewDefaultLogger(RootPath, Cfg.LogFile)
	Logger.SetLogConsole(true)
	if err != nil {
		Logger.SetMaxFileCount(5)
		Logger.SetMaxFileSize(10, fileLogger.MB)
		Logger.SetLogLevel(fileLogger.LEVEL(1))
		Logger.E(err.Error())
		Buff = make([]byte, BUFFERSIZE)
	} else {
		Logger.SetMaxFileCount(Cfg.LogCount)
		Logger.SetMaxFileSize(int64(Cfg.LogSize), fileLogger.MB)
		Logger.SetLogLevel(fileLogger.LEVEL(Cfg.LogLevel))
		Buff = make([]byte, Cfg.Buffer)
	}
}

func NotValidTime() bool {
	var nowTime = time.Now().Format("15:04")
	if Cfg.From < Cfg.To {
		return !(nowTime >= Cfg.From && nowTime <= Cfg.To)
	} else if Cfg.From == Cfg.To {
		return false
	} else if Cfg.From > Cfg.To {
		return !(nowTime >= Cfg.From && nowTime <= "23:59" || nowTime >= "00:00" && nowTime <= Cfg.To)
	}
	return false
}

func (s FtpEntryList) Len() int {
	return len(s)
}

func (s FtpEntryList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// order by file modified time desc
func (s FtpEntryList) Less(i, j int) bool {
	return s[i].Time.Unix() < s[j].Time.Unix()
}

func (s FileInfoList) Len() int {
	return len(s)
}

func (s FileInfoList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// order by file modified time desc
func (s FileInfoList) Less(i, j int) bool {
	return s[i].ModTime().Unix() < s[j].ModTime().Unix()
}
