# 文件同步工具 / File synchronization tool
Synchronize local files with ftp/sftp server.
Or synchronize remote files with localhost.

_周期性地上传本地文件到远端ftp/sftp服务器，或从远端ftp/sftp下载到本地_


### 功能/Features：
* run as Windows backend service
* _作为Windows服务后台运行_
* resume from break-point
* _断点续传_
* ftp and sftp suported
* _支持ftp和sftp_
* download or upload mode
* _上传或下载模式切换_
* work time setting
* _可设置工作时段_

### configure.ini
```ini
[filesync]
protocol=ftp            # [ftp | sftp]
host=127.0.0.1          # server address
port=21                 # port
user=test               # username
passwd=123456           # password
path=E:\\data           # local path
remote_path=/           # remote path
interval=20             # minutes
retry=3                 # retry N times while failure occurs
log_file=filesync.log
log_level=0
log_size=10
log_count=5
new_file_first=true     # latest modified file come first
timeout=10              # seconds
serv_name=FileSync      # installing as a windows service
work_during=00:00-23:59 # 22:00-04:00 is also OK   
buffer=4096             # buffer size
trans_mode=download     # [upload | download]
```
## Windows:
### 安装/Install

_命令行下运行_ / In command line：
```cmd
filesync.exe install
sc start filesync
```
或双击 install.cmd / Or double click install.cmd

![image](https://github.com/wn0112/filesync/assets/14155504/c4984d83-bb3d-4481-a61d-46071cf55d01)
![image](https://github.com/wn0112/filesync/assets/14155504/76c9ba08-88c7-4460-b4c9-21cc53c16881)

### 删除/Uninstall
```cmd
sc stop filesync
sc delete filesync
```
或双击 uninstall.cmd / Or double click uninstall.cmd
