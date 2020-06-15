cd %~dp0

sc stop ftpupload
sc delete ftpupload
sc stop ftp_client
sc delete ftp_client
filesync.exe install
sc start filesync
