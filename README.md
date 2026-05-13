# 企业资质证书与许可证到期管理

技术栈：Go（Gin + GORM）+ SQLite（纯 Go 驱动 [glebarez/sqlite](https://github.com/glebarez/sqlite) 基于 modernc.org/sqlite）后端，React 18 + Vite + Ant Design 前端，单进程在 **8080** 同时提供 `/api` 与静态资源；**单镜像多阶段构建**，SQLite 数据文件建议挂载到 **`/app/data/certs.db`**。

## 功能概要

- **证书台账**：类别涵盖营业执照、行业许可证、认证证书、安全生产许可、特种经营许可、资质等级证书；状态按到期日与一级提醒阈值计算「有效 / 即将到期 / 已过期」，支持手动「已注销」；支持扫描件与正反面 URL、CSV 按类别导出。
- **提醒配置**（`/api/reminders`）：每类 180/90/30 天三级窗口与默认负责人；台账中「即将到期」按**一级窗口**（默认 180 天）判定，若需以 90 天为界可将该类一级阈值改为 90。
- **续期申请**：材料清单（JSON）、预计提交日、进展（材料准备中→已提交→审核中→已通过）；通过后可将说明写回证书备注（PATCH `new_cert_notes`）。
- **附件与年检**：证书维度 URL；历届年检记录表；前端 Drawer 查看。
- **费用**：行政/代理费用（以**分**存储）；按年汇总与按类别对比。
- **合规视图**：仪表盘数量卡片、按类统计、90 天内紧迫度排序、年度到期日历（按月汇总）、到期待办（即将到期/已过期自动同步）。
- **预置数据**：`init.sql` 在**证书表为空**时导入约 **100** 条演示证书及近三年年检、续期、费用样例。

## 本地开发

**依赖**：Go **1.22+**、Node **18+**。

```powershell
# 后端（默认 SQLite: .\data\certs.db ，预置脚本 .\init.sql）
cd web
npm install
npm run build
cd ..
go run .
```

前端开发可另开端口并由 Vite 代理 API：

```powershell
cd web
npm run dev
```

此时后端需单独 `go run .`（8080），浏览器访问 `http://localhost:5173`。

## 测试

```powershell
go test ./...
```

## Docker 单容器

构建：

```powershell
docker build -t license-expiry:latest .
```

运行（与 `prompt.txt` 一致，数据卷持久化）：

```powershell
docker run --rm -d -p 8080:8080 -v certs-data:/app/data license-expiry:latest
```

访问 `http://localhost:8080`。首次启动且库为空时会执行 `/app/init.sql` 种子数据。

环境变量：

| 变量 | 含义 | 默认 |
|------|------|------|
| `SQLITE_PATH` | SQLite 文件路径 | `./data/certs.db`（镜像内 `/app/data/certs.db`） |
| `INIT_SQL_PATH` | 预置 SQL 路径 | `./init.sql`（镜像内 `/app/init.sql`） |
| `PORT` | 监听端口（仅数字，不含冒号） | `8080` |

验证镜像后若曾使用 Compose，请执行与测试相同的 compose 项目目录下的 `docker compose down -v` 进行清理；仅用 `docker run` 时可用 `docker stop <container>` 停止容器。

## API 摘要

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/health` | 健康检查 |
| GET | `/api/dashboard` | 合规总览 |
| GET | `/api/certificates` | 列表，`?category=` |
| GET | `/api/certificates/urgent` | 90 天内紧迫排序 |
| GET | `/api/certificates/export?category=` | CSV（UTF-8 BOM） |
| GET/POST/PUT/DELETE | `/api/certificates/:id` | CRUD |
| GET/PUT | `/api/reminders`、`/api/reminders/:id` | 提醒配置 |
| GET/POST/PATCH | `/api/renewals`、`:id` | 续期申请 |
| GET/POST | `/api/inspections` | 年检记录 |
| GET/POST | `/api/fees` | 费用明细 |
| GET | `/api/fees/summary?year=` | 年度汇总 |
| GET | `/api/fees/by-category?year=` | 按类对比 |
| GET | `/api/calendar/:year` | 年度日历数据 |
| GET/PATCH | `/api/todos`、`/api/todos/:id` | 待办 |

## 目录结构

- `main.go` — 入口、嵌入 `web/dist`
- `internal/` — 模型、数据库、HTTP 处理、状态计算
- `web/` — 前端源码；`web/dist` 为构建产物（提交前可自行构建）
- `init.sql` — 空库时预置数据
- `Dockerfile` — 前端构建 + Go 编译 + Alpine 运行镜像
