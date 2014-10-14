unsafessh
=========
[![Gobuild Download](http://gobuild.io/badge/github.com/codeskyblue/unsafessh/downloads.svg)](http://gobuild.io/github.com/codeskyblue/unsafessh)

exec remote command easily.

developing

## Usage
run in client
```
unsafessh --addr=unix:/tmp/unsafessh.sock exec -- echo hello
```

run in server
```
unsafessh --addr=unix:/tmp/unsafessh.sock serv
```
