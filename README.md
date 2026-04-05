# mysqlrouter_exporter

A productized Prometheus exporter for MySQL Router.

This project is intended to be usable outside [`dbbot`](https://github.com/fanderchan/dbbot), while remaining compatible with [`dbbot`](https://github.com/fanderchan/dbbot)'s existing YAML config format and service launch style.

Chinese README: [README.zh-CN.md](README.zh-CN.md)

## What It Collects

- Router availability and scrape health
- Router build info and start time
- Per-route active connections, total connections, blocked hosts, health, and destinations
- Metadata refresh counters and last refresh endpoint info
- Optional route connection byte counters
- Local listener checks by parsing `mysqlrouter.conf`

The listener checks are the main differentiator of this implementation: they verify that the ports declared in `mysqlrouter.conf` are actually listening.

## Install

### Build locally

```bash
cd /usr/local/mysqlrouter_exporter
bash ./build.sh
```

Artifacts:

- `build/mysqlrouter_exporter`
- `dist/mysqlrouter_exporter-<goos>-<goarch>`

### Install on a host

```bash
cd /usr/local/mysqlrouter_exporter
bash ./scripts/install.sh
```

Then edit `/etc/mysqlrouter_exporter/config.yml` and start:

```bash
systemctl daemon-reload
systemctl enable --now mysqlrouter_exporter
```

## Run Modes

### 1. Legacy-compatible YAML config

```bash
mysqlrouter_exporter --config /etc/mysqlrouter_exporter/config.yml
```

### 2. Env vars

```bash
export MYSQLROUTER_EXPORTER_ROUTER_API_BASE_URL="https://127.0.0.1:8443/api/20190715"
export MYSQLROUTER_EXPORTER_ROUTER_API_USER="router_api_user"
export MYSQLROUTER_EXPORTER_ROUTER_API_PASSWORD_FILE="/etc/mysqlrouter_exporter/router_api_password"
export MYSQLROUTER_EXPORTER_WEB_LISTEN_ADDRESS=":9165"
mysqlrouter_exporter
```

### 3. Flags

```bash
mysqlrouter_exporter \
  --router.api-base-url https://127.0.0.1:8443/api/20190715 \
  --router.api-user router_api_user \
  --router.api-password-file /etc/mysqlrouter_exporter/router_api_password \
  --web.listen-address :9165
```

## Compatibility Flags And Env Vars

To reduce migration friction from other Router exporters, this project also accepts:

- Flags: `--url`, `--user`, `--pass`, `--listen-port`, `--skip-tls-verify`, `--tls-ca-cert-path`, `--tls-cert-path`, `--tls-key-path`
- Env vars: `MYSQLROUTER_EXPORTER_URL`, `MYSQLROUTER_EXPORTER_USER`, `MYSQLROUTER_EXPORTER_PASS`, `MYSQLROUTER_TLS_CACERT_PATH`, `MYSQLROUTER_TLS_CERT_PATH`, `MYSQLROUTER_TLS_KEY_PATH`

## Configuration

Example:

```yaml
listen_address: ":9165"
metrics_path: "/metrics"
api_base_url: "https://127.0.0.1:8443/api/20190715"
api_user: "router_api_user"
api_password_file: "/etc/mysqlrouter_exporter/router_api_password"
timeout_seconds: 5
insecure_skip_verify: true
tls_ca_file: ""
tls_cert_file: ""
tls_key_file: ""
collect_route_connections: false
router_config_file: "/var/lib/mysqlrouter/mysqlrouter.conf"
listener_check_enabled: true
listener_check_timeout_seconds: 1
```

Notes:

- `api_password` and `api_password_file` are both supported.
- `collect_route_connections` is off by default to reduce scrape cost and cardinality pressure.
- `listener_check_enabled` parses `mysqlrouter.conf` and performs TCP checks against configured listeners.

## Endpoints

- Metrics: `/metrics`
- Health: `/-/healthy`
- Ready: `/-/ready`

## Release

This repository includes:

- `build.sh` for local builds
- `.github/workflows/ci.yml` for test and build verification
- `.github/workflows/release.yml` for tag-based GitHub releases
- `.goreleaser.yml` for portable release artifacts
- `Dockerfile` for container packaging

Recommended tag flow:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## [`dbbot`](https://github.com/fanderchan/dbbot) Compatibility

[`dbbot`](https://github.com/fanderchan/dbbot) can continue to deploy `mysqlrouter_exporter` as before.

The intended compatibility path is:

- build this standalone project under `/usr/local/mysqlrouter_exporter`
- let [`dbbot`](https://github.com/fanderchan/dbbot) prefer the standalone artifact when present
- keep [`dbbot`](https://github.com/fanderchan/dbbot)'s bundled binary as a fallback

That means [`dbbot`](https://github.com/fanderchan/dbbot) integration remains usable even if this standalone project is not built yet.

## License

Apache-2.0
