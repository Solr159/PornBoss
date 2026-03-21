# Pornboss 后端

一个使用 Go、Gin、GORM 与 SQLite 实现的本地视频文件管理后台。启动时会根据配置文件中的目录同步视频文件，支持标签管理，并提供简单的 HTTP 接口进行查询与播放。

## 快速开始

```bash
go run ./cmd/server -addr :8080
```

启动后可访问：

- `GET /healthz` 健康检查
- `GET /videos?limit=100&offset=0` 列出视频
- `GET /videos?tags=tag1,tag2` 根据标签交集分页查询
- `GET /videos/{id}` 查看单个视频
- `GET /videos/{id}/stream` 直接播放视频流
- `GET /videos/{id}/thumbnail` 获取缩略图
- `GET /tags` / `POST /tags` / `PATCH /tags/{id}` / `DELETE /tags/{id}` 管理标签
- `POST /videos/tags/add` / `POST /videos/tags/remove` 传入 `{"video_ids":[...],"tag_id":123}` 批量为视频新增或删除指定标签
- `POST /sync` 手动触发全量同步https://github.com/JavBoss/pornboss/pull/new/complete_package

## 同步与标签

- 启动时自动扫描配置中的目录，将识别出的常见视频格式记录到 SQLite。
- 每个文件会基于抽样 MD5 指纹，在重命名或移动后仍能识别同一个视频并更新记录。
- 同步时会刷新文件大小、修改时间，并删除不存在的旧记录。
- 通过 `POST /sync` 可以手动重新扫描。
- 标签支持新增、删除、重命名，并可对多个视频批量增删标签，查询接口支持按标签交集过滤。
- 首次发现新视频或检测到文件变化时会自动生成缩略图，缩略图文件以视频指纹命名并存放在 `thumbnails_dir` 指定的目录下。

## 运行测试

```bash
GOCACHE=$(pwd)/.gocache go test ./...
```

## 注意事项

- 默认使用 GORM + `gorm.io/driver/sqlite`（基于 `github.com/mattn/go-sqlite3`），需要 CGO 与 C 编译链。
- 通过交互式 CLI 下载 ffmpeg/ffprobe 到 `bin/<platform>/`（例如 `bin/windows-x86_64`）。若下载的平台等于当前平台，会同步到 `internal/bin/` 供运行时使用（可通过环境变量 `FFMPEG_PATH` 覆盖）。
- 数据库文件会自动存放到 `database_path` 指定的位置，缩略图存放在 `thumbnails_dir`。
- 数据库存储的 `path` 为相对于配置根目录的相对路径，配合 `directory` 字段可还原真实路径。

## Web 前端（React + Tailwind + Zustand）

已在 `web/` 目录集成前端界面，提供视频网格列表、缩略图、分页、标签筛选与批量为视频添加/移除标签，点击视频弹出模态框播放（不跳转）。

开发与自检：

```bash
# 进入前端目录并安装依赖
cd web
npm install

# 本地开发（已配置 Vite 代理到后端 http://localhost:8080）
npm run dev

# 代码规范检查（ESLint）
npm run lint

# 生产构建输出到 web/dist
npm run build
```

注意：请先启动后端（默认 :8080），前端开发服务器会通过代理访问后端接口，避免 CORS 问题。

## 开发脚本（交互式 CLI）

使用 Node + Inquirer 的交互式命令行整合 dev/release/download 功能（需先打包生成可发布脚本）：

```bash
# 首次打包 CLI
cd scripts/cli
npm install
npm run build

# 进入交互菜单（scripts/cli.sh 会在缺少产物时自动构建）
scripts/cli.sh

# 直接启动后端/前端
scripts/cli.sh dev backend
scripts/cli.sh dev frontend

# 下载 ffmpeg/ffprobe
scripts/cli.sh download linux-x86_64

# 可选环境变量
#   ADDR=:8080           后端监听端口
#   WITH_STATIC=1        后端同时托管已构建的前端（指向 STATIC 指定目录，默认 web/dist）
#   STATIC=web/dist      托管的前端目录
#   SKIP_NPM_INSTALL=1   启动前端时跳过 npm install
```

## 生产部署（后端托管前端）

- 构建前端静态资源：

```bash
cd web && npm run build
```

- 启动后端并托管前端：

```bash
go run ./cmd/server -addr :8080 -static web/dist
```

- 说明：
  - 后端参数 `-static` 指向前端构建目录（默认 `web/dist`）。
  - 后端自动托管 `/assets` 与根路径 `/`，并提供 SPA fallback（非 API 路由返回 `index.html`）。
  - 相对路径配置均相对于当前工作目录解析，适合打包后的独立运行。

## 打包发布

- 通过交互式 CLI 打包（release）：

```bash
# 进入交互菜单 → release → 选择平台 → 输入版本号
scripts/cli.sh

# 或直接指定平台与版本号
scripts/cli.sh release linux-x86_64 v0.1.0
scripts/cli.sh release windows-x86_64 v0.1.0

# 可选环境变量：
#   SKIP_WEB_BUILD=1  跳过前端构建（使用现有 web/dist）
```

- 发布包内容：
  - `pornboss` 或 `pornboss.exe`（后端可执行文件）
  - `web/dist`（前端静态产物）
  - `internal/bin/ffmpeg(.exe)`、`internal/bin/ffprobe(.exe)`

发布包为 `release/pornboss-<version>-<platform>.zip`，解压后运行：

```bash
# macOS / Linux
./pornboss -addr :8080 -static web/dist

# Windows
pornboss.exe -addr :8080 -static web/dist
```

注意事项：
  - SQLite 依赖 CGO，跨平台构建需准备相应的交叉编译工具链；Windows 目标需要可用的 MinGW（例如 `x86_64-w64-mingw32-gcc`）。
  - ffmpeg/ffprobe 从 `bin/<platform>/` 打包进入发布包（并复制到包内的 `internal/bin/`）；未找到则回退系统 PATH 或 `FFMPEG_PATH` / `FFPROBE_PATH`。
