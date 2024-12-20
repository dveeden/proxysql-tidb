WARNING: This is for proof-of-concept demonstration only.

This can be scheduled via the [ProxySQL Scheduler](https://proxysql.com/documentation/scheduler/) like this:

```
INSERT INTO scheduler(id, active, interval_ms, filename) values (1, 1, 10000, '/tmp/proxysql-tidb');
LOAD SCHEDULER TO RUNTIME;
SAVE SCHEDULER TO DISK;
```

You then should check `journalctl -u proxysql -f` or elswhere for output.

What this does is to do a soft offline of TiDB servers that are in a graceful shutdown.

This require the TiDB server to have a non-zero [`graceful-wait-before-shutdown`](https://docs.pingcap.com/tidb/stable/tidb-configuration-file#graceful-wait-before-shutdown-new-in-v50)

Possible future improvements:
- Logging via [slog](https://pkg.go.dev/log/slog)
- More defensive coding and error checking
- Also onlineing servers when they are healty again
- Adding hostgroup filtering

Other possible solutions:
- [pingcap/tidb#58008](https://github.com/pingcap/tidb/pull/58008): Let TiDB no longer respond with OK on `COM_PING` during shudown.
- Add a global variable for the server health/shutdown status. See [pingcap/tidb#58007](https://github.com/pingcap/tidb/issues/58007).
