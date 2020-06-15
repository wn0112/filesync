cd %~dp0

sc stop ftpupload
sc delete ftpupload
sc stop ftp_client
sc delete ftp_client
file_trans.exe install
sc start file_trans