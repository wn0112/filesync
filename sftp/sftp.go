package sftp

import (
	"fmt"
	. "ftp_upload/global"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Sftp struct {
	Conn		*sftp.Client
	Host		string
	User		string
	Pass		string
	Port		int
}


func (f *Sftp) Connect() error {
	var retryTimes = 0

	RETRY:
		retryTimes += 1

		sshCfg := ssh.ClientConfig {
			User: f.User,
			Auth: []ssh.AuthMethod{ssh.Password(f.Pass)},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
			Timeout: time.Duration(Cfg.Timeout) * time.Second,
		}

		// 建立 ssh 连接
		sshCon, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", f.Host, f.Port), &sshCfg)
		if err != nil {
			if retryTimes < Cfg.Retry {
				Logger.I("Retrying to connect: [ %s:%d ]", f.Host, f.Port)
				time.Sleep(1 * time.Second)
				goto RETRY
			} else {
				return err
			}
		}

		// 建立 sftp 连接
		f.Conn, err = sftp.NewClient(sshCon)
		if err != nil {
			if retryTimes < Cfg.Retry {
				Logger.I("Retrying to connect: [ %s:%d ]", f.Host, f.Port)
				goto RETRY
			} else {
				return err
			}
		}
	return nil
}


func (f *Sftp) Close() error {
	if f.Conn != nil {
		err := f.Conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}


func (f *Sftp)UploadDirectory(localPath string, remotePath string) error {
	absPath, _ := filepath.Abs(localPath)
	Logger.I("=== Now in local path [ %s ]", absPath)
	existedFiles := make(map[string]int64)
	var (
		fullLocalFilePath string
		fullLocalPath string
		fullRemotePath string
		offset int64
		retryTimes = 0
	)

	// 远端目录不存在则创建
	_ = f.Conn.Mkdir(remotePath)

	// 获取远端文件夹内文件列表
	RETRY:
		retryTimes += 1
		entries, err := f.Conn.ReadDir(remotePath)
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
		if !entry.IsDir() {
			existedFiles[entry.Name()] = int64(entry.Size())
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
			Logger.I("Stop tag is [ %v ], Not valid time is [ %v ]", Stop,NotValidTime())
			break
		}

		if !item.IsDir() {
			offset = 0
			remoteFileSize, exist := existedFiles[item.Name()]
			if exist {
				// 文件大小不同, 断点续传
				if item.Size() > remoteFileSize {
					offset = remoteFileSize
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
			_ = f.Conn.Mkdir(fullRemotePath)
			err = f.UploadDirectory(fullLocalPath, fullRemotePath)
			if err != nil {
				Logger.E("Error: %s -> %s", err.Error(), fullLocalPath)
			}
		}
	}

	return nil
}


func (f *Sftp) UploadFile(remotePath string, fullFilePath string, offset int64) error {
	var (
		retryTimes = 0
		fullRemoteFilePath string
		N int
		readErr error
	)

	Logger.I("Uploading file:\t%s", fullFilePath)
	fp, err := os.Open(fullFilePath)
	if err != nil {
		return err
	}
	_, _ = fp.Seek(offset, 0)
	defer fp.Close()

	RETRY:
		retryTimes += 1
		fullRemoteFilePath = fmt.Sprintf("/%s/%s",
			strings.TrimLeft(remotePath, "/"),
			filepath.Base(fullFilePath))
		remoteFp, err := f.Conn.OpenFile(fullRemoteFilePath, os.O_CREATE | os.O_WRONLY | os.O_APPEND)
		if err != nil {
			if retryTimes < Cfg.Retry {
				Logger.E(err.Error())
				time.Sleep(1 * time.Second)
				goto RETRY
			} else {
				return err
			}
		}

		// 指针移动到文件末尾
	_, _ = remoteFp.Seek(0, 2)
	defer remoteFp.Close()

	for {
		N, readErr = fp.Read(Buff)
		if readErr != nil && readErr != io.EOF {
			return readErr
		}

		_, err = remoteFp.Write(Buff[:N])
		if err != nil {
			return err
		}

		if readErr == io.EOF {
			break
		}
	}
	TotalFile += 1
	return nil
}


func (f *Sftp) DownloadDirectory(localPath string, remotePath string) error {
	Logger.I("=== Now in remote path [ %s ]", remotePath)
	existedFiles := make(map[string]int64)
	var (
		fullRemoteFilePath string
	    fullLocalPath string
		fullRemotePath string
		offset int64
		retryTimes = 0
	)

	// 获取远端文件夹内文件列表
	RETRY:
		retryTimes += 1
		entries, err := f.Conn.ReadDir(remotePath)
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
		sort.Sort(sort.Reverse(FileInfoList(entries)))
	} else {
		sort.Sort(FileInfoList(entries))
	}

	// 下载根目录所有文件
	for _, entry := range entries {
		if Stop || NotValidTime() {
			Logger.I("Stop tag is [ %v ], Not valid time is [ %v ]", Stop, NotValidTime())
			break
		}

		if !entry.IsDir() {
			offset = 0
			localFileSize, exist := existedFiles[entry.Name()]
			if exist {
				// 文件大小不同, 断点续传
				if entry.Size() > localFileSize {
					offset = localFileSize
				} else {
					continue
				}
			}

			fullRemoteFilePath = "/" + strings.TrimLeft(fmt.Sprintf("%s/%s", remotePath, entry.Name()), "/")
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
		if entry.IsDir() {
			fullLocalPath = filepath.Join(localPath, entry.Name())
			fullRemotePath = "/" + strings.TrimLeft(fmt.Sprintf("%s/%s", remotePath, entry.Name()), "/")
			_ = os.Mkdir(fullLocalPath, os.ModePerm)
			err = f.DownloadDirectory(fullLocalPath, fullRemotePath)
			if err != nil {
				Logger.E("Error: %s -> %s", err.Error(), fullRemotePath)
			}
		}
	}

	return nil
}


func (f *Sftp) DownloadFile(localPath string, fullRemoteFilePath string, offset int64) error {
	var (
		retryTimes = 0
		N = 0
		readErr error
	)

	Logger.I("Downloading file:\t%s", fullRemoteFilePath)
	fp, err := os.OpenFile(filepath.Join(localPath, filepath.Base(fullRemoteFilePath)),
		os.O_CREATE | os.O_WRONLY | os.O_APPEND, os.ModePerm)

	if err != nil {
		return err
	}
	defer fp.Close()

	// 移动指针到文件末尾
	_, err = fp.Seek(0, 2)
	if err != nil {
		return err
	}

	RETRY:
		retryTimes += 1
		remoteFp, err := f.Conn.Open(fullRemoteFilePath)
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

	defer remoteFp.Close()
	// 断点续传, 从offset 处开始接收
	_, err = remoteFp.Seek(offset, 0)
	if err != nil {
		Logger.E(err.Error())
		return err
	}

	for {
		N, readErr = remoteFp.Read(Buff)
		if readErr != nil && readErr != io.EOF {
			return readErr
		}

		_, err = fp.Write(Buff[:N])
		if err != nil {
			return err
		}

		if readErr == io.EOF {
			break
		}
	}

	TotalFile += 1
	return nil
}
