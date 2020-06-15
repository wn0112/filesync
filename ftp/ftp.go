package ftp

import (
	"fmt"
	. "ftp_upload/global"
	"github.com/jlaffaye/ftp"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var	err error

type Ftp struct {
	Conn		*ftp.ServerConn
	Host		string
	User		string
	Pass		string
	Port		int
}

func (f *Ftp) Connect() error {
	var retryTimes = 0

	RETRY:
		retryTimes += 1
		f.Conn, err = ftp.Dial(fmt.Sprintf("%s:%d", f.Host, f.Port),
			ftp.DialWithTimeout(time.Duration(Cfg.Timeout) * time.Second),
			ftp.DialWithLocation(time.Local))

		if err != nil {
			if retryTimes < Cfg.Retry {
				Logger.I("Retrying to connect: [ %s:%d ]", f.Host, f.Port)
				time.Sleep(1 * time.Second)
				goto RETRY
			} else {
				return err
			}
		}

	// 登录
	err = f.Conn.Login(f.User, f.Pass)
	if err != nil {
		return err
	}
	return nil
}


func (f *Ftp) Close() error {
	if f.Conn != nil {
		_ = f.Conn.Logout()
		err := f.Conn.Quit()
		return err
	}
	return nil
}


func (f *Ftp)UploadDirectory(localPath string, remotePath string) error {
	absPath, _ := filepath.Abs(localPath)
	Logger.I("=== Now in local path [ %s ]", absPath)
	existedFiles := make(map[string]int64)
	var fullLocalFilePath string
	var fullLocalPath string
	var fullRemotePath string
	var offset uint64
	var retryTimes = 0

	// 远端目录不存在则创建
	_ = f.Conn.MakeDir(remotePath)

	// 获取远端文件夹内文件列表
	RETRY:
		retryTimes += 1
		entries, err := f.Conn.List(remotePath)
		if err != nil {
			// list file 有时出错，重试
			if retryTimes < Cfg.Retry {
				Logger.I("Retrying list: [ %s ]", remotePath)
				time.Sleep(1 * time.Second)
				goto RETRY
			} else {
				return err
			}
		}

	// 创建远端已存在文件映射
	for _, entry := range entries {
		if entry.Type == ftp.EntryTypeFile {
			existedFiles[entry.Name] = int64(entry.Size)
		}
	}

	// 获取本地文件夹内文件列表
	items, err := ioutil.ReadDir(localPath)
	if err != nil {
		return err
	}

	// 排序
	if Cfg.Reverse {
		sort.Sort(sort.Reverse(FileInfoList(items)))
	} else {
		sort.Sort(FileInfoList(items))
	}

	// 上传当前文件夹所有文件
	for _, item := range items {
		if Stop || NotValidTime() {
			Logger.I("Stop tag is [ %v ], Not valid time is [ %v ]", Stop, NotValidTime())
			break
		}

		if !item.IsDir() {
			offset = 0
			remoteFileSize, exist := existedFiles[item.Name()]
			if exist {
				// 文件大小不同, 断点续传
				if item.Size() > remoteFileSize {
					offset = uint64(remoteFileSize)
				} else {
					continue
				}
			}

			fullLocalFilePath = filepath.Join(localPath, item.Name())
			err = f.UploadFile(remotePath, fullLocalFilePath, offset)
			if err != nil {
				Logger.E("Error: %s -> %s", err.Error(), filepath.Join(remotePath, filepath.Base(fullLocalFilePath)))
			}
		}
	}

	// 上传当前文件夹所有子文件夹
	for _, item := range items {
		if Stop || NotValidTime() {
			Logger.I("Stop tag is [ %v ], Not valid time is [ %v ]", Stop, NotValidTime())
			break
		}
		if item.IsDir() {
			fullLocalPath = filepath.Join(localPath, item.Name())
			fullRemotePath = "/" + strings.TrimLeft(fmt.Sprintf("%s/%s", remotePath, item.Name()), "/")
			_ = f.Conn.MakeDir(fullRemotePath)
			err = f.UploadDirectory(fullLocalPath, fullRemotePath)
			if err != nil {
				Logger.E("Error: %s -> %s", err.Error(), fullLocalPath)
			}
		}
	}

	return nil
}


func (f *Ftp) UploadFile(remotePath string, fullFilePath string, offset uint64) error {
	var retryTimes = 0
	Logger.I("Uploading file:\t%s", fullFilePath)
	fp, err := os.Open(fullFilePath)
	if err != nil {
		return err
	}
	fp.Seek(int64(offset), 0)
	defer fp.Close()

	RETRY:
		retryTimes += 1
		err = f.Conn.StorFrom(fmt.Sprintf("/%s/%s", strings.TrimLeft(remotePath, "/"),
			filepath.Base(fullFilePath)), fp, offset)
		if err != nil {
			// 失败重试 N 次
			if retryTimes < Cfg.Retry {
				Logger.I("Retrying file:\t\t%s", fullFilePath)
				time.Sleep(1 * time.Second)
				goto RETRY
			} else {
				return err
			}
		}
	TotalFile += 1
	return nil
}


func (f *Ftp) DownloadDirectory(localPath string, remotePath string) error {
	Logger.I("=== Now in remote path [ %s ]", remotePath)
	existedFiles := make(map[string]int64)
	var fullRemoteFilePath string
	var fullLocalPath string
	var fullRemotePath string
	var offset int64
	var retryTimes = 0

	// 获取远端文件夹内文件列表
	RETRY:
		retryTimes += 1
		entries, err := f.Conn.List(remotePath)
		if err != nil {
			// list file 有时出错，重试
			if retryTimes < Cfg.Retry {
				Logger.I("Retrying list: [ %s ]", remotePath)
				time.Sleep(1 * time.Second)
				goto RETRY
			} else {
				return err
			}
		}

	// 获取本地文件夹内文件列表
	items, err := ioutil.ReadDir(localPath)
	if err != nil {
		return err
	}

	// 创建本地已存在文件映射
	for _, item := range items {
		if !item.IsDir() {
			existedFiles[item.Name()] = item.Size()
		}
	}
	// 排序
	if Cfg.Reverse {
		sort.Sort(sort.Reverse(FtpEntryList(entries)))
	} else {
		sort.Sort(FtpEntryList(entries))
	}

	// 下载根目录所有文件
	for _, entry := range entries {
		if Stop || NotValidTime() {
			Logger.I("Stop tag is [ %v ], Not valid time is [ %v ]", Stop, NotValidTime())
			break
		}
		if entry.Type == ftp.EntryTypeFile {
			offset = 0
			localFileSize, exist := existedFiles[entry.Name]
			if exist {
				// 文件大小不同, 断点续传
				if entry.Size > uint64(localFileSize) {
					offset = localFileSize
				} else {
					continue
				}
			}

			fullRemoteFilePath = "/" + strings.TrimLeft(fmt.Sprintf("%s/%s", remotePath, entry.Name), "/")
			err = f.DownloadFile(localPath, fullRemoteFilePath, offset)
			if err != nil {
				Logger.E("Error: %s -> %s", err.Error(), fullRemoteFilePath)
			}
		}
	}

	// 下载根目录所有子文件夹
	for _, entry := range entries {
		if Stop || NotValidTime() {
			Logger.I("Stop tag is [ %v ], Not valid time is [ %v ]", Stop, NotValidTime())
			break
		}
		if entry.Type == ftp.EntryTypeFolder {
			fullLocalPath = filepath.Join(localPath, entry.Name)
			fullRemotePath = "/" + strings.TrimLeft(fmt.Sprintf("%s/%s", remotePath, entry.Name), "/")
			_ = os.Mkdir(fullLocalPath, os.ModePerm)
			err = f.DownloadDirectory(fullLocalPath, fullRemotePath)
			if err != nil {
				Logger.E("Error: %s -> %s", err.Error(), fullRemotePath)
			}
		}
	}

	return nil
}


func (f *Ftp) DownloadFile(localPath string, fullRemoteFilePath string, offset int64) error {
	var retryTimes = 0
	var N = 0

	Logger.I("Downloading file:\t%s", fullRemoteFilePath)
	fp, err := os.OpenFile(filepath.Join(localPath, filepath.Base(fullRemoteFilePath)),
		os.O_CREATE | os.O_WRONLY | os.O_APPEND, os.ModePerm)

	if err != nil {
		return err
	}
	defer fp.Close()

	// 移动指针到末尾
	_, err = fp.Seek(0, 2)
	if err != nil {
		return err
	}

	RETRY:
		retryTimes += 1
		// 断点续传, 从offset 处开始接收
		resp, err := f.Conn.RetrFrom(fullRemoteFilePath, uint64(offset))
		if err != nil {
			// 失败重试 N 次
			if retryTimes < Cfg.Retry {
				time.Sleep(1 * time.Second)
				Logger.I("Retrying file:\t\t%s", fullRemoteFilePath)
				goto RETRY
			} else {
				return err
			}
		}

	defer resp.Close()

	for {
		N, err = resp.Read(Buff)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		_, err = fp.Write(Buff[:N])
		if err != nil {
			return err
		}
	}

	TotalFile += 1
	return nil
}

