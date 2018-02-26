chat-server
===========
Learning go by implementing a classic programming problem and working my
way haphazardly through [Go by Example](https://gobyexample.com/) :).

I'm certain this is not idiomatic go, but someday it will be. I'm also certain
this is suboptimal and buggy as hell, but I'm going for trial-by-fire learning
here. Someday I'll turn this into a blog post...

to use
======

Build the server and start it

```
$  go build chatserver.go
$  ./chatserver
```

Open 1 or more terminals and connect using `telnet`:

```
$ telnet localhost 8080
Trying ::1...
Connected to localhost.
Escape character is '^]'.
hello!
[::1]:55360: hello!
echoooo
[::1]:55360: echoooo
```

