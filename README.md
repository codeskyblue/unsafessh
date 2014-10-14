unsafessh
=========

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
