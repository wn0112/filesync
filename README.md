# file sync tool
# `文件同步工具`
Sync local files to ftp/sftp server at regular intervals.
Or sync remote files to localhost.

`周期性地上传本地文件到远端ftp/sftp服务器，或从远端ftp/sftp下载到本地`

Features:

`功能`：
* run as Windows backend service
* `作为Windows服务后台运行`
* resume from break-point
* `断点续传`
* ftp and sftp suported
* `支持ftp和sftp`
* download or upload mode
* `上传或下载模式切换`
* work time setting
* `可设置工作时段`

### config.ini
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
### Install
### `安装`
In command line:

`命令行下运行：`
```cmd
filesync.exe install
sc start filesync
```
### Uninstall
### `删除`
```cmd
sc stop filesync
sc delete filesync
```
