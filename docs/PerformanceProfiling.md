# Performance Profiling
Projects support built in performance monitoring and debugging via the pprof tools. To run projects with profiling enabled, ensure the ENABLE_PPROF=1 environment variable is set.

## golang webserver
(https://go.dev/blog/pprof)
### Install Graphviz
Ensure you have an updated version of Graphviz(https://graphviz.org/about/) installed for visualizing profile outputs.

```
apt install -y graphviz
```
### Collect a Profile
1. Start projects with profiling enabled: `ENABLE_PPROF=1 go run ./webserver`
2. Collect a Profile in desired format (e.g. png): `go tool pprof -png -seconds=10 http://127.0.0.1:80/debug/pprof/allocs?seconds=10 > .pprof/allocs.png`
    a. Replace “allocs” with the name of the profile(https://pkg.go.dev/runtime/pprof#Profile) to collect.
    b. Replace the value of seconds with the amount of time you need to reproduce performance issues.
    c. Read more about the available profiling URL parameters here(https://pkg.go.dev/net/http/pprof#hdr-Parameters).
    d. `go tool pprof` does not need to run on the same host as the project, just ensure you provide the correct HTTP url in the command. Note that Graphviz must be installed on the system you’re running `pprof` from.
3. Reproduce any interactions with the project that you’d like to collect profiling information for.
4. A graph visualization of the requested performance profile should now be saved locally, take a look and see what’s going on.

## rust
cargo bench --bench fibonacci_bench --profile-time 12345