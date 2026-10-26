# AI 一键部署实现规范

新项目可说：按本文实现一键部署。

## 目标

```
本机：docker build --platform linux/amd64 → docker save → 发布包.tar.gz
         │ 人工上传
服务器：解压 → 编辑配置文件 `.env`  → docker load → docker compose up -d → 配现有 Nginx
```

默认目标服务器平台为 **`linux/amd64`**（常见 x86_64 Linux）。本机若是 Apple Silicon（arm64），不指定平台会打出 `linux/arm64` 镜像，服务器上会出现 platform 不匹配。

## 硬性要求

1. 服务器已有 Nginx — 不起 Nginx 容器；反代片段写在 `docs/deploy.md` 即可。
2. 防冲突 — 容器/卷用项目前缀；依赖不占 `3306`/`6379`；应用只绑 `127.0.0.1:WEB_PORT`。
3. 源码不进服务器 — Compose **无 `build:`**；一键运行命令是 `docker compose up -d`。
4. Nginx 配好即可访问 — 页面与 API 尽量同源。
5. 配置可维护 — 运行配置只放配置文件 `.env`。
6. 必须有运维文档 `docs/deploy.md`。
7. **平台对齐 `linux/amd64`** —
   - 本机打包：`docker build --platform linux/amd64 -t <image> .`
   - `docker-compose.yml` 中应用服务写明 `platform: linux/amd64`
   - `docs/deploy.md` 必须写清该参数，并收录 arm64/amd64 不匹配的排错
   - 服务器若是 arm64，再改为 `linux/arm64`（须与服务器一致，且文档写明）

## 必须交付（4 + 文档）

| 文件 | 作用 |
| --- | --- |
| `Dockerfile` | 本机构建 |
| `docker-compose.yml` | 服务器 `up -d`（仅 `image:`，含 `platform`） |
| `.env.example` | 配置模板 |
| `docs/deploy.md` | 打包命令（含 `--platform`）、`docker load`、`up -d`、Nginx 片段、运维 |

打包可用文档里的几条 `docker` 命令；**不要求**再拆 `deploy/*.sh` 或单独的 `nginx-location.conf`。

发布包内容：`*-web.tar.gz` + `docker-compose.yml` + `.env.example`（+ 可选基础镜像 tar）。`docs/deploy.md` 可一并打进包，或运维单独留一份。

## 服务器标准步骤

```bash
cp .env.example .env          # 改密码、端口
gzip -dc *-web.tar.gz | docker load
docker compose up -d          # 一键运行
# 按 docs/deploy.md 把 location 贴进现有 Nginx
```

## 禁止

- 服务器默认 `--build` / 默认私有仓库 `pull`
- 容器裸名 `mysql`、`redis`
- 再起抢 80/443 的 Nginx
- 缺少 `docs/deploy.md`、提交 `.env`
- 默认按本机 arch 构建却部署到 amd64 服务器（漏写 `--platform linux/amd64`）