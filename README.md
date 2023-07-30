# useful-code-snippets-in-golang
This repository  contains some real world code snippets into golang. Those come from some of my mutiple go-based projects.
feel free to learn from them. For any help regarding this repo - contact me at https://blog.cloudmentor-scale.com/contact

* ### [dual-servers-http-https-with-custom-tls.go](https://github.com/jeamon/useful-code-snippets-in-golang/blob/master/dual-servers-http-https-with-custom-tls.go)

This is a simple go program that demonstrates how to spin up HTTP & HTTPS servers at the same time
and force all HTTP traffic to be redirected to HTTPS while configurting HSTS. All done with customized 
TLS configurations. You will find some useful openssl commands to generate your self-signed certificates.
This program expects to load key and self-signed certificate from child directory named "certificates".

*commands examples to generate certificates with OpenSSL 1.1.1c*
```
[+] RSA-Based keys - Self Signed Certificate Generation (key - recommendation key ≥ 2048) 

1/ openssl req -x509 -nodes -newkey rsa:4096 -keyout server.rsa.key -out server.rsa.crt -days 365

[+] ECDSA-Based keys - Self Signed Certificate Generation (key - recommendation key ≥ secp384r1)**

1/ To list ECDSA supported curves use following command : openssl ecparam -list_curves  

2/ openssl req -new -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -x509 -nodes -days 365 -out server.ecdsa.crt -keyout server.ecdsa.key  

3/ openssl req -x509 -nodes -newkey ec:secp384r1 -keyout server.ecdsa.key -out server.ecdsa.crt -days 365  

4/ openssl req -x509 -nodes -newkey ec:<(openssl ecparam -name secp384r1) -keyout server.ecdsa.key -out server.ecdsa.crt -days 365  

TLS   : Transport Layer Security
ECDHE : Ephemeral Elliptic Curve Diffie-Hellman
RSA   : Rivest–Shamir–Adleman
ECDSA : Elliptic Curve Digital Signature Algorithm
GCM   : Galois Counter Mode
SHA   : Secure Hash Algorithm

```

* ### [demo-native-http-server.go](https://github.com/jeamon/useful-code-snippets-in-golang/blob/master/demo-native-http-server.go)

This is another go web development program that constructs a HTTP server with a middleware in charge of generating & assigning an ID
to each request that comes in. It shows up the useage of the built-in shutdown feature (available from 1.8+). This is achieved
with two separate goroutines (one which handles exit system calls - another waiting for a notification in order to call the shutdown)
and a timeout-based context. Finally, you'll find a simple way to associate a customized logger to the server. Below are some outputs.

```

[+] Outputs on server console.

[ 2:11:23] {nxos-geek}:~$ go run http-web.go
2021/08/26 02:15:45 web server is starting ...
2021/08/26 02:15:48 received new request  [id: 9c4565b075528b7f] - logging ...
[request] 2021/08/26 02:15:48 [id: 9c4565b075528b7f] [ip: 127.0.0.1:63448] [method: GET] [url: /] [browser: curl/7.55.1]
2021/08/26 02:15:48 completed new request [id: 9c4565b075528b7f] - thanks ...
2021/08/26 02:15:52 received new request  [id: b56b0ea8c72dfcd6] - logging ...
[request] 2021/08/26 02:15:52 [id: b56b0ea8c72dfcd6] [ip: 127.0.0.1:63451] [method: GET] [url: /welcome] [browser: curl/7.55.1]
2021/08/26 02:15:52 completed new request [id: b56b0ea8c72dfcd6] - thanks ...
2021/08/26 02:15:57 received new request  [id: 7f9e1905baf425b6] - logging ...
[request] 2021/08/26 02:15:57 [id: 7f9e1905baf425b6] [ip: 127.0.0.1:63461] [method: GET] [url: /] [browser: curl/7.55.1]
2021/08/26 02:15:57 completed new request [id: 7f9e1905baf425b6] - thanks ...
2021/08/26 02:16:01 received new request  [id: 17ba4ecf3741bfb0] - logging ...
[request] 2021/08/26 02:16:01 [id: 17ba4ecf3741bfb0] [ip: 127.0.0.1:63467] [method: GET] [url: /welcome] [browser: curl/7.55.1]
2021/08/26 02:16:01 completed new request [id: 17ba4ecf3741bfb0] - thanks ...
received signal type:  interrupt
exit command received. exiting ...
2021/08/26 02:16:05 shutting down the web server ... please wait for 45 secs max

[ 2:16:05] {nxos-geek}:~$


[+] Outputs on client console.

[ 2:11:18] {nxos-geek}:~$ curl localhost:8080/
404 page not found

[ 2:15:48] {nxos-geek}:~$ curl localhost:8080/welcome
simple plain welcome.
[ 2:15:52] {nxos-geek}:~$ curl localhost:8080/
404 page not found

[ 2:15:57] {nxos-geek}:~$ curl localhost:8080/welcome
simple plain welcome.
[ 2:16:01] {nxos-geek}:~$

``` 

* ### [contextual-command-executor.go](https://github.com/jeamon/useful-code-snippets-in-golang/blob/master/multiple-commands-executor/main.go)

This cross-platform tool allows to execute multiple commands from shell with possibility to specify single execution timeout for all of the commands.
Each command output will be streaming to a unique file named with command id (order number into the list) suffixed by the hour minutes and seconds of 
program start time. Be aware that all these output files will be saved into a unique daily folder. If the daily folder doesn't not exist it will be created.
In event of creation failure, the program aborts it execution. The timeout and execution cancellation is achieved with builtin CommandContext feature available 
from go v1.7. Please find below some details on how to use this tool.

```
Usage:
    
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

```


* ### [dynamic-webapp-spam-loader.go](https://github.com/jeamon/useful-code-snippets-in-golang/blob/master/auto-spam-words-loader/main.go)

This is a short routine that runs in background as goroutine every customized hours which will load a list of your defined spam words and 
store them into the program memory for usage. It loads the spam file content only if the size of the file has changed or if the latest
modification date attribute changed. This means we keep track of these two values. Updating the in-memory list store uses RW mutex to
facilitates concurent reading access. There is a code snippet which demonstrates how you can use this spam loader routine into your code.
The idea is simple. Each time a user submit a message, you need to check if the message contains one of your spam words before proceding. 

```go

// contact submission message format.
type Message struct {
	FullName string
	Email string
	Subject string
	Content string
	Errors map[string]string
}

// check if subject or content is suspicious.
func (msg *Message) isSpamMessage() bool {
	// acquire the lock and ensure its release.
	addSpamMutex.RLock()
	defer addSpamMutex.RUnlock()
	// loop over the spam words list and stop once there is a hit
	for _, spam := range spamWords {
		if strings.Contains(msg.Content, spam) || strings.Contains(msg.Subject, spam) {
			return true
		}
	}

	return false
}

// snipped to insert into the contact handler and send fake confirmation for spam message.
if msg.isSpamMessage() {
	// routine to handle goes here
	// you can ignore user message or send fake confirmation
}

```