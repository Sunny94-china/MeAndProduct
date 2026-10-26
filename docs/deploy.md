# MeAndProduct 运维部署

按 [AI 一键部署规范](./superpowers/ai-one-click-deploy-spec.md) 实施。

## 架构约定

| 项 | 说明 |
| --- | --- |
| 业务容器 | `meandproduct-backend` / `meandproduct-public` / `meandproduct-admin` |
| 平台 | 默认 **`linux/amd64`**（常见 x86_64 服务器） |
| MySQL / Redis | 使用**服务器已有**实例，Compose **不起**这两个容器 |
| Nginx | 使用**服务器已有** Nginx，不起 Nginx 容器，不占 80/443 |
| 源码 | **不进服务器**；Compose 无 `build:`，服务器命令是 `docker compose up -d` |
| 配置 | 只维护 `.env`（从 `.env.example` 复制） |

本机监听（默认，仅 `127.0.0.1`）：

- API：`18080`
- 前台 + 工作台：`18081`
- 管理后台：`18082`

## 一、本机打包（有源码）

目标平台默认 `linux/amd64`。在 Apple Silicon（arm64）上**必须**指定平台，否则服务器会报 platform 不匹配。

### 方式 A：一条 Makefile（推荐）

```bash
# 在仓库根目录
make pack TAG=v1.0.0
# 产出：dist/meandproduct-v1.0.0.tar.gz
```

### 方式 B：手写 docker 命令

```bash
export TAG=v1.0.0
export BACKEND_IMAGE=meandproduct-backend:${TAG}
export PUBLIC_IMAGE=meandproduct-public:${TAG}
export ADMIN_IMAGE=meandproduct-admin:${TAG}
export DOCKER_DEFAULT_PLATFORM=linux/amd64

docker compose -f docker-compose.yml -f docker-compose.build.yml build

mkdir -p dist/release
docker save "${BACKEND_IMAGE}" "${PUBLIC_IMAGE}" "${ADMIN_IMAGE}" \
  | gzip > "dist/release/meandproduct-${TAG}-web.tar.gz"

# 发布目录（无源码）
mkdir -p "dist/pack/meandproduct"
cp docker-compose.yml .env.example docs/deploy.md "dist/pack/meandproduct/"
cp "dist/release/meandproduct-${TAG}-web.tar.gz" "dist/pack/meandproduct/"
# 按 TAG 写好镜像名后，也可再编辑一份 .env.example 打进包
```

等价的单镜像构建示例：

```bash
docker build --platform linux/amd64 -t meandproduct-backend:v1.0.0 ./backend
docker build --platform linux/amd64 -t meandproduct-public:v1.0.0 ./frontend/public-site
docker build --platform linux/amd64 -t meandproduct-admin:v1.0.0 ./frontend/admin-dashboard
```

发布包内应有：

- `meandproduct-<TAG>-web.tar.gz`（三个业务镜像）
- `docker-compose.yml`
- `.env.example`
- `deploy.md`（本文，可选但推荐）

**不要**把 `backend/`、`frontend/`、`docker-compose.build.yml` 拷到服务器。

## 二、服务器准备

### 1. MySQL（已有实例）

```sql
CREATE DATABASE IF NOT EXISTS meandproduct
  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'meandproduct'@'%' IDENTIFIED BY '换成强密码';
GRANT ALL ON meandproduct.* TO 'meandproduct'@'%';
FLUSH PRIVILEGES;
```

库名 / 用户使用 `meandproduct` 前缀，避免和其它业务冲突。

### 2. Redis（已有实例）

建议使用独立 DB 序号（默认 `.env` 里 `REDIS_DB=11`），不要占用其它应用的 `0` 号库。

### 3. 容器访问宿主机上的 MySQL / Redis

Compose 已配置 `host.docker.internal:host-gateway`。若连不上：

- 确认 MySQL `bind-address` 允许 Docker 网桥访问（或监听 `0.0.0.0` 并由防火墙限制）
- `.env` 里 `MYSQL_HOST` / `REDIS_ADDR` 使用 `host.docker.internal`

## 三、服务器部署（无源码）

```bash
# 1) 上传并解压发布包（不要 git clone 整仓）
tar -xzf meandproduct-v1.0.0.tar.gz
cd meandproduct

# 2) 配置
cp .env.example .env
# 编辑 .env：MYSQL_*、REDIS_*、JWT_SECRET、ADMIN_PASSWORD、端口等

# 3) 加载镜像 + 启动
gzip -dc meandproduct-*-web.tar.gz | docker load
docker compose up -d

# 4) 健康检查
curl -sS "http://127.0.0.1:${API_PORT:-18080}/api/v1/health"
```

常用运维：

```bash
docker compose ps
docker compose logs -f backend
docker compose down          # 停业务容器；不动宿主机 MySQL/Redis
```

## 四、接入服务器已有 Nginx（同源）

页面与 API 走同一域名，浏览器无跨域问题。把下面片段并入现有 `server { ... }`，或单独放到 `/etc/nginx/conf.d/meandproduct.conf`。

将 `18080` / `18081` / `18082` 换成你在 `.env` 里配置的 `API_PORT` / `PUBLIC_PORT` / `ADMIN_PORT`。

```nginx
upstream meandproduct_api {
    server 127.0.0.1:18080;
    keepalive 16;
}
upstream meandproduct_public {
    server 127.0.0.1:18081;
    keepalive 8;
}
upstream meandproduct_admin {
    server 127.0.0.1:18082;
    keepalive 8;
}

server {
    listen 80;
    # listen 443 ssl;
    # ssl_certificate     /etc/nginx/ssl/meandproduct.crt;
    # ssl_certificate_key /etc/nginx/ssl/meandproduct.key;
    server_name meandproduct.example.com;

    client_max_body_size 20m;

    location /api/ {
        proxy_pass http://meandproduct_api;
        proxy_http_version 1.1;
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection        "";
        proxy_read_timeout 60s;
    }

    # 末尾斜线去掉 /admin 前缀，管理端容器按站点根路径提供资源
    location /admin/ {
        proxy_pass http://meandproduct_admin/;
        proxy_http_version 1.1;
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection        "";
    }

    location = /admin {
        return 301 /admin/;
    }

    location = /founder {
        return 302 /studio;
    }
    location /founder/ {
        return 302 /studio;
    }

    location / {
        proxy_pass http://meandproduct_public;
        proxy_http_version 1.1;
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection        "";
    }
}
```

```bash
nginx -t && nginx -s reload
```

访问：

- 前台 / 工作台：`https://你的域名/`
- 管理后台：`https://你的域名/admin/`
- API：`https://你的域名/api/v1/...`

## 五、`.env` 关键字段

| 字段 | 说明 |
| --- | --- |
| `TAG` / `*_IMAGE` | 与打包时镜像名一致 |
| `API_PORT` / `PUBLIC_PORT` / `ADMIN_PORT` | 仅本机回环；不要用 80/443 |
| `MYSQL_*` | 宿主机已有 MySQL；库/用户建议 `meandproduct` |
| `REDIS_ADDR` / `REDIS_PASSWORD` / `REDIS_DB` | 宿主机 Redis；`REDIS_DB` 建议 `11` |
| `JWT_SECRET` | 生产必改，`openssl rand -hex 32` |
| `ADMIN_*` | 启动时写入/更新的管理员账号 |

`.env` **不要**提交到 Git。

## 六、排错

### 1. `The requested image's platform (linux/arm64) does not match ... (linux/amd64)`

本机（多为 Apple Silicon）打包时漏了平台。重新打包：

```bash
make pack TAG=v1.0.0
# 内部使用 DOCKER_DEFAULT_PLATFORM=linux/amd64 / compose platform: linux/amd64
```

或手写：

```bash
docker build --platform linux/amd64 -t meandproduct-backend:v1.0.0 ./backend
```

服务器若是 **arm64**，需改为 `linux/arm64`：本机构建加 `--platform linux/arm64`，并把 `docker-compose.yml` 里三处 `platform:` 改成 `linux/arm64` 后重新打包。

### 2. 健康检查失败 / 连不上数据库

```bash
docker compose logs backend
# 检查 .env 的 MYSQL_HOST、密码、库是否已创建
# 宿主机测试：mysql -h127.0.0.1 -umeandproduct -p meandproduct
```

### 3. 管理后台静态资源 404

确认 Nginx `location /admin/` 的 `proxy_pass` **带尾斜线**（见上文），以便去掉 `/admin` 前缀。

### 4. 端口被占用

```bash
ss -lntp | grep -E '18080|18081|18082'
```

改 `.env` 端口后，同步改 Nginx upstream，再 `docker compose up -d` 与 `nginx -s reload`。

## 七、禁止事项（对照规范）

- 服务器上默认 `docker compose up -d --build`
- 服务器默认依赖私有仓库 `docker compose pull`（本流程以 `docker load` 为准）
- 容器裸名 `mysql` / `redis`，或再起抢 80/443 的 Nginx
- 把源码仓库整份放到服务器
- 提交 `.env`
- 本机按默认 arch 构建却部署到 amd64 服务器（漏写 `linux/amd64`）
