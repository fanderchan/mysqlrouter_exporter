# mysqlrouter_exporter

MySQL Router 的产品化 Prometheus exporter。

这个项目的目标是在脱离 [`dbbot`](https://github.com/fanderchan/dbbot) 的场景下也能独立使用，同时继续兼容 [`dbbot`](https://github.com/fanderchan/dbbot) 现有的 YAML 配置格式和服务启动方式。

English README: [README.md](README.md)

## 采集内容

- Router 可用性和抓取健康状态
- Router 构建信息和启动时间
- 每条路由的活跃连接数、总连接数、阻塞主机数、健康状态和目的端
- 元数据刷新计数器和最近一次刷新端点信息
- 可选的路由连接字节计数
- 通过解析 `mysqlrouter.conf` 执行本地监听端口检查

本实现的主要差异点在于监听检查：它会验证 `mysqlrouter.conf` 中声明的端口是否真的处于监听状态。

## 安装

### 本地构建

```bash
cd /usr/local/mysqlrouter_exporter
bash ./build.sh
```

产物：

- `build/mysqlrouter_exporter`
- `dist/mysqlrouter_exporter-<goos>-<goarch>`

### 安装到目标主机

```bash
cd /usr/local/mysqlrouter_exporter
bash ./scripts/install.sh
```

然后编辑 `/etc/mysqlrouter_exporter/config.yml` 并启动：

```bash
systemctl daemon-reload
systemctl enable --now mysqlrouter_exporter
```

## 运行模式

### 1. 兼容旧版的 YAML 配置

```bash
mysqlrouter_exporter --config /etc/mysqlrouter_exporter/config.yml
```

### 2. 环境变量

```bash
export MYSQLROUTER_EXPORTER_ROUTER_API_BASE_URL="https://127.0.0.1:8443/api/20190715"
export MYSQLROUTER_EXPORTER_ROUTER_API_USER="router_api_user"
export MYSQLROUTER_EXPORTER_ROUTER_API_PASSWORD_FILE="/etc/mysqlrouter_exporter/router_api_password"
export MYSQLROUTER_EXPORTER_WEB_LISTEN_ADDRESS=":9165"
mysqlrouter_exporter
```

### 3. 命令行参数

```bash
mysqlrouter_exporter \
  --router.api-base-url https://127.0.0.1:8443/api/20190715 \
  --router.api-user router_api_user \
  --router.api-password-file /etc/mysqlrouter_exporter/router_api_password \
  --web.listen-address :9165
```

## 兼容参数与环境变量

为降低从其他 Router exporter 迁移的成本，这个项目还兼容以下输入：

- 参数：`--url`、`--user`、`--pass`、`--listen-port`、`--skip-tls-verify`、`--tls-ca-cert-path`、`--tls-cert-path`、`--tls-key-path`
- 环境变量：`MYSQLROUTER_EXPORTER_URL`、`MYSQLROUTER_EXPORTER_USER`、`MYSQLROUTER_EXPORTER_PASS`、`MYSQLROUTER_TLS_CACERT_PATH`、`MYSQLROUTER_TLS_CERT_PATH`、`MYSQLROUTER_TLS_KEY_PATH`

## 配置

示例：

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

说明：

- 同时支持 `api_password` 和 `api_password_file`
- `collect_route_connections` 默认关闭，用于降低抓取成本和基数压力
- `listener_check_enabled` 会解析 `mysqlrouter.conf` 并对配置的监听地址执行 TCP 检查

## 接口

- 指标：`/metrics`
- 健康检查：`/-/healthy`
- 就绪检查：`/-/ready`

## 发布

仓库内包含：

- `build.sh`：本地构建
- `.github/workflows/ci.yml`：测试和构建校验
- `.github/workflows/release.yml`：基于 tag 的 GitHub Release
- `.goreleaser.yml`：可移植发布产物
- `Dockerfile`：容器镜像打包

推荐的打 tag 流程：

```bash
git tag v0.1.0
git push origin v0.1.0
```

## License

Apache-2.0
