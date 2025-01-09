# Streamer for sql dump loading (PostgersQL)

PROcessing Dump & Loading

Example:

```bash
go run main.go load -c 'pg://user:password@127.0.0.1:5432/database' ./postgresdb.backup.gz
```

For installation in linux, just run helper script with "install" command:

```bash
./buildew install
```
