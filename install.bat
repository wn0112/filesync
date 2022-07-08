cd %~dp0

sc stop filesync
sc delete filesync
filesync.exe install
sc start filesync
