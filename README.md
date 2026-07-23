# XiuGo

**XiuGo** `5.0.0` — a Go rewrite of Xiuno BBS 4.x, built with GoFrame v2.

Production runs **Go only**. Optional local PHP tree is reference-only.

## Quick start

```sh
cp manifest/config/config.example.yaml manifest/config/config.yaml
# edit database + paths
./start-local.sh
```

- Site: http://127.0.0.1:8081
- Admin: http://127.0.0.1:8081/admin

Branches: `develop` (active work), `main` (stable). Version: **5.0.0**.

See `docs/` for migration notes.
