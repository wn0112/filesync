# file sync tool
Sync local files to ftp/sftp server in cycle.
Or sync remote files to local.

### config.ini
```ini
[params]
protocol=ftp    # [ftp | sftp]
host=127.0.0.1  # server address
port=21         # port
user=test       # username
passwd=123456   # password
path=E:\data    # local path
remote_path=/   # remote path
interval=20     # minutes
retry=3         # retry N times while failure occurs
log_level=0
log_size=10
log_count=5
new_file_first=true   # latest modified file come first
timeout=10
serv_name=File_trans  # installing as a windows service
work_during=00:00-23:59
buffer=4096           # buffer size
trans_mode=down       # [upload | download]
```
### Install
In command line:
```cmd
file_trans.exe install
sc start file_trans
```
### Uninstall
```cmd
sc stop file_trans
sc delete file_trans
```
