package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// this cross-platform tool allows to execute multiple commands from shell with possibility
// to specify single execution timeout for all of the commands. Each command output will be
// streaming to a unique file named with command id (order number into the list) suffixed
// by the hour minutes and seconds of program start time. Be aware that all these output files
// will be saved into a unique daily folder. If the daily folder doesn't not exist it will
// be created. In event of creation failure, the program aborts it execution. The timeout and
// execution cancellation is achieved with builtin CommandContext feature available from go v1.7.

// Version  : 1.0
// Author   : Jerome AMON
// Created  : 24 August 2021

// list of tasks.
var tasks []string

// set list of commands to be executed based on platform (windows or linux).
func setDefaultTasks() {
	// command syntax for windows platform.
	if runtime.GOOS == "windows" {
		tasks = []string{"systeminfo", "tasklist", "netstat -n 5", "ping 8.8.8.8 -t", "ipconfig /all"}
	} else {
		tasks = []string{"hostnamectl", "lscpu && lsmem && lsusb && lspci && lshw && lsblk", "uname -r", "df -h",
			"sysinfo", "du | less", "/bin/ps -aux", "netstat -c 5", "ping 8.8.8.8", "tail -f /var/log/messages"}
	}
}

func formatSyntax(cmd *exec.Cmd, task string, ctx context.Context) *exec.Cmd {
	// command syntax for windows platform.
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", task)
	} else {
		// set default shell to use on linux.
		shell := "/bin/sh"
		// load shell name from env variable.
		if os.Getenv("SHELL") != "" {
			shell = os.Getenv("SHELL")
		}
		// syntax for linux-based platforms.
		cmd = exec.CommandContext(ctx, shell, "-c", task)
	}

	return cmd
}

// handlesignal is a function that process SIGTERM from kill command or CTRL-C or more.
func handlesignal(cancel context.CancelFunc) {
	// one signal to be handled.
	sigch := make(chan os.Signal, 1)
	// setup supported exit signals.
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL,
		syscall.SIGTERM, syscall.SIGHUP, os.Interrupt, os.Kill)

	// block until something comes in.
	<-sigch
	// close the signal channel to not receive any more.
	signal.Stop(sigch)
	// call context cancellation so all goroutines got notified to exit.
	cancel()
	return
}

func executeTask(i int, cmd *exec.Cmd, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	var err error
	// this start the task asynchronously.
	err = cmd.Start()
	if err != nil {
		// failed to start the task. no need to continue
		log.Printf("task [%02d] failed to start - errmsg : %v\n", i, err)
		return
	}
	log.Printf("task [%02d] execution successfully started under process id [%d]\n", i, cmd.Process.Pid)
	// goroutine to handle the blocking behavior of wait func - channel used to notify.
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	// watch on both channels and handle the case which hits/triggers first.
	select {

	case <-ctx.Done():
		// timeout reached - or context cancel called.
		switch ctx.Err() {
		case context.DeadlineExceeded:
			// timeout reached.
			log.Printf("task [%02d] execution timeout reached - killing the process id [%d]\n", i, cmd.Process.Pid)
		case context.Canceled:
			// context cancellation called.
			log.Printf("task [%02d] execution cancellation requested - killing the process id [%d]\n", i, cmd.Process.Pid)
		}

		// kill the process and exit from this function.
		if err = cmd.Process.Kill(); err != nil {
			log.Printf("task [%02d] execution - failed to kill process id [%d] - errmsg: %v\n", i, cmd.Process.Pid, err)
		} else {
			log.Printf("task [%02d] execution - succeeded to kill process id [%d]\n", i, cmd.Process.Pid)
		}

		return
	case err = <-done:
		// task execution completed [cmd.wait func] - check if for error.
		if err != nil {
			log.Printf("task [%02d] execution completed under process id [%d] with failure - errmsg : %v\n", i, cmd.Process.Pid, err)
		} else {
			log.Printf("task [%02d] execution completed under process id [%d] with success\n", i, cmd.Process.Pid)
		}
		return
	}

}

// createFolder ensures daily folder exists if not then creates it and returns
// if the program should exit or continue. if needs to continue sends folder name.
func createFolder(starttime time.Time) (bool, string) {
	// format daily folder suffix based on startup time.
	logtime := fmt.Sprintf("%d%02d%02d", starttime.Year(), starttime.Month(), starttime.Day())
	outfolder := fmt.Sprintf("outputs-%s", logtime)
	err := os.Mkdir(outfolder, 0755)

	if err == nil {
		// folder sucessfully created.
		return false, outfolder
	}

	if os.IsExist(err) {
		// there is a file or folder with the name.
		info, err := os.Stat(outfolder)
		if err != nil {
			// failed to retrieve details. stop program.
			return true, ""
		}

		if !info.IsDir() {
			// there is a file using the name. stop program
			fmt.Printf(" [-] Program aborted. failed to create the outputs folder. name already in use.")
			return true, ""
		}
		// folder already exists so continue program.
		return false, outfolder
	}
	// abort program for other error.
	return true, ""
}

func main() {
	// get startup time.
	starttime := time.Now()

	// will be triggered to display usage instructions.
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n", usage)
	}

	timeoutPtr := flag.Int("timeout", 0, "commands execution timetout value in seconds")

	// check for any valid subcommands : version or help
	if len(os.Args) == 2 {
		if os.Args[1] == "version" || os.Args[1] == "--version" || os.Args[1] == "-v" {
			fmt.Fprintf(os.Stderr, "\n%s\n", version)
			os.Exit(0)
		}

		if os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h" {
			fmt.Fprintf(os.Stderr, "\n%s\n", usage)
			os.Exit(0)
		}
	}

	flag.Parse()

	if *timeoutPtr <= 0 {
		// reset to default 180 secs.
		*timeoutPtr = 180
		log.Printf("all tasks will be executed with a defaultt imeout value of %d secs\n", *timeoutPtr)
	} else {
		log.Printf("all tasks will be executed with a timeout value of %d secs\n", *timeoutPtr)
	}

	// considering all others arguments provided after -timeout / --timeout flag as tasks.
	tasksList := flag.Args()
	if len(tasksList) == 0 {
		log.Printf("no tasks provided for execution - default demo %q tasks will be used\n", runtime.GOOS)
		setDefaultTasks()
	} else {
		tasks = tasksList
		log.Printf("detected [%d] tasks to execute with a timeout value of %d secs\n", len(tasks), *timeoutPtr)
	}

	var wg sync.WaitGroup
	// ctx, cancel := context.WithCancel(context.Background())

	// create context with wanted timeout value - 10 secs
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutPtr)*time.Second)
	// call cancel at end to release resources.
	defer cancel()
	// goroutine to handle user/os signal like CTRL-C.
	go handlesignal(cancel)

	// create dedicated daily outputs folder.
	shouldExit, outfolder := createFolder(starttime)
	if shouldExit {
		return
	}

	// preformat each output file suffix.
	filesuffix := fmt.Sprintf("%02d%02d%02d.txt", starttime.Hour(), starttime.Minute(), starttime.Second())

	var cmd *exec.Cmd
	// iterate over tasks list.
	for i, task := range tasks {
		// construct output file path based on task number from local working directory.
		filepath := fmt.Sprintf("%s%s%d.%s", outfolder, string(os.PathSeparator), i, filesuffix)
		file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			// lets user know on console and continue on next task.
			log.Printf("task [%02d] failed to create or open output file - errmsg : %v\n", i, err)
			continue
		}
		// set file as destination for combined outputs.
		// cmd.Stdout, cmd.Stderr = file, file
		// prepare task according to OS.
		cmd = formatSyntax(cmd, task, ctx)
		// set file as destination for combined outputs.
		cmd.Stdout, cmd.Stderr = file, file
		// increment due to each goroutine.
		wg.Add(1)
		// launch task execution asynchronously.
		go executeTask(i, cmd, ctx, &wg)
	}
	// block until all tasks done or exit.
	wg.Wait()
}

const version = "This tool is <commands-executor> • version 1.0 By Jerome AMON"

const usage = `Usage:
    
    Please in case you want to define a timeout value - it should come just after the program name
    and before all the tasks command to be executed. Notice that these tasks should be double quoted.
    Be aware that for demonstration purpose, if no tasks are provided - default tasks will be used.
    
    command-executor [-timeout <execution-deadline-in-seconds>] "task-one" "task-two" . . . "task-three"

    Examples:

    --- on Windows

    go run contextual-command-executor.go "tasklist" "ipconfig /all" "systeminfo"
    go run contextual-command-executor.go -timeout 120 "tasklist" "ipconfig /all" "systeminfo"

    --- on Linux

    go run contextual-command-executor.go "ps" "ifconfig" "cat /proc/cpuinfo" "ping 8.8.8.8"
    go run contextual-command-executor.go -timeout 120 "ps" "ifconfig" "cat /proc/cpuinfo" "ping 8.8.8.8"

    --- help or version checkings

    go run contextual-command-executor.go --help
    go run contextual-command-executor.go --version
    go run contextual-command-executor.go -h
    go run contextual-command-executor.go -v

`
