unsafessh
=========
[![Gobuild Download](http://gobuild.io/badge/github.com/codeskyblue/unsafessh/downloads.svg)](http://gobuild.io/github.com/codeskyblue/unsafessh)

exec remote command easily.

* no password
* stream stdin,stdout,stderr
* same exitcode in server and client
* send client signal(INT,HUP,TERM) to server

## Usage
run in server
```
unsafessh serv
```

run in client
```
unsafessh exec -- python -i
```

