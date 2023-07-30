@echo off
set d=%DATE% && set t=%TIME:~0,-3% && set bt=%d:~0,-1%%t%
Rem git rev-parse --short HEAD > tempFile && set /P gitCommit=<tempFile
git rev-list -1 HEAD > tempFile && set /P gitCommit=<tempFile
go env GOVERSION > tempFile && set /P goVersion=<tempFile
go env GOOS > tempFile && set /P targetOS=<tempFile
go env GOARCH > tempFile && set /P targetArch=<tempFile

del tempFile

go build -o demo-build-flags.exe -a -ldflags "-X 'main.BuildTime=%bt%' -X 'main.GitCommit=%gitCommit%' -X 'main.APIVersion=1.0.0' -X 'main.WebVersion=1.0.0' -X 'main.Version=1.0.0' -X 'main.TargetOS=%targetOS%' -X 'main.TargetArch=%targetArch%' -X 'main.GoVersion=%goVersion%'" demo-build-flags.go