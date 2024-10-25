# Debugging the finch daemon

This document outlines where to find/access logs and how to configure profiling tools for finch daemon.

## Logs

Logs are the first place to check when you suspect a problem with finch-daemon. If `finch-daemon` was started via `systemd` then you can obtain logs using `journalctl`:

```shell
sudo journalctl -u finch
```

> **Note**
> The command above assumes that you have used the unit file definition [finch.service](../finch.service) we have provided. If you have created your own unit file for `finch-daemon` and replace `finch-daemon` with the one you have made. Amazon Linux distributions of Finch also use the name `finch` for the finch-daemon service.

If you have started `finch-daemon` manually, logs will either be emitted to stderr/stdout.

## CPU Profiling

We can use Golangs `pprof` tool to profile the daemon. To enable profiling you must set the `--debug-addr` CLI parameter when invoking `finch-daemon`:

```shell
./finch-daemon --debug-addr localhost:6060
```

> **Note**
> Similarly to adding the command line option for a local run of finch-daemon, any systemd service file can also be modified to include the `--debug-addr` option.


Once you have configured the debug address you can send a `GET` to the `/debug/pprof/profile` endpoint to receive a CPU profile of the daemon. You can specify an optional argument `seconds` to limit the results to a certain time span:

```shell
curl http://localhost:6060/debug/pprof/profile?seconds=40 > out.pprof
```

You can use the `pprof` tool provided by the Go CLI to visualize the data within a web browser:

```shell
go tool pprof -http=:8080 out.pprof
```

For more information on pprof, [see its documentation here](https://pkg.go.dev/net/http/pprof).
