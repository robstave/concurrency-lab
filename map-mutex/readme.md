Mutex-protected map example
===========================

to run
```bash

go run ./cmd/...
```


This directory contains a simple `map[int]int` wrapped with a `sync.RWMutex` to
make concurrent access safe.

Package
-------
- `package RegularIntMap` defines `RegularIntMap` with methods:
	- `Load(key int) (int, bool)`
	- `Store(key, value int)`
	- `Delete(key int)`

Demo
----
`cmd/regular-map-demo/main.go` shows basic usage: concurrent stores, loads, and a delete.

Run
---
From the module folder (`map-mutex/`) or repo root:

```
go run ./map-mutex/cmd/regular-map-demo
```

You should see printed key/value pairs and confirmation of a deleted key.



notes

 https://medium.com/@deckarep/the-new-kid-in-town-gos-sync-map-de24a6bf7c2c
 