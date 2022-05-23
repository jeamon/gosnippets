package main

import (
	"fmt"
	"strings"
)

var (
	Version    string
	Author     string = "Jerome Amon <cloudmentor.scale@gmail.com>"
	BuildTime  string
	GitCommit  string
	GoVersion  string
	TargetOS   string
	TargetArch string
	WebVersion string
	APIVersion string
	SourceLink string = "https://github.com/jeamon/web-based-jobs-worker-service/commit/"
)

func main() {
	fmt.Println("Version:", Version)
	fmt.Println(" Web version:", WebVersion)
	fmt.Println(" API version:", APIVersion)
	fmt.Println(" Go version:", strings.TrimLeft(GoVersion, "go"))
	fmt.Println(" Build Time:", BuildTime)
	fmt.Println(" OS/Arch:", TargetOS+"/"+TargetArch)
	fmt.Println(" Git commit:", GitCommit)
	fmt.Println(" Author:", Author)
	fmt.Println(" Source:", SourceLink+GitCommit)
}

// %variable:~startposition,numberofchars%
// %variable:~num_chars_to_skip,numberofchars%

// set gover=go env GOVERSION
// set GOHOSTARCH=amd64
// set GOHOSTOS=windows
// set d=%DATE% && set t=%TIME:~0,-3% && set bt="%d:~0,-1% at %t:~0,-1%" && set bt=%bt:~1,-1%
/*
@echo off
set d=%DATE% && set t=%TIME:~0,-3% && set bt=%d:~0,-1%%t%
Rem git rev-parse --short HEAD > tempFile && set /P gitCommit=<tempFile
git rev-list -1 HEAD > tempFile && set /P gitCommit=<tempFile
go env GOVERSION > tempFile && set /P goVersion=<tempFile
go env GOOS > tempFile && set /P targetOS=<tempFile
go env GOARCH > tempFile && set /P targetArch=<tempFile

del tempFile

go build -o build-flags.exe -a -ldflags "-X 'main.BuildTime=%bt%' -X 'main.GitCommit=%gitCommit%' -X 'main.APIVersion=1.0.0' -X 'main.WebVersion=1.0.0' -X 'main.Version=1.0.0' -X 'main.TargetOS=%targetOS%' -X 'main.TargetArch=%targetArch%' -X 'main.GoVersion=%goVersion%'" build-flags.go
*/
