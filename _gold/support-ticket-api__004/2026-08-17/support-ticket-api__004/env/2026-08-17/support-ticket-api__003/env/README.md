# support-ticket-api

小型工单系统 API，使用 Gin 和 PostgreSQL 实现，支持客户、客服、主管三种角色。

## 运行环境

- Go 1.26
- PostgreSQL，默认连接 `postgres://postgres:postgres@127.0.0.1:55432/support_ticket?sslmode=disable`
- 默认 HTTP 端口 `18106`

## 初始化

```bash
go mod tidy
go run ./cmd/migrate
go run ./cmd/api
```

也可以通过环境变量覆盖默认配置：

```bash
export HTTP_PORT=18106
export DATABASE_URL='postgres://postgres:postgres@127.0.0.1:55432/support_ticket?sslmode=disable'
```

## 测试

```bash
go test ./...
```

## 角色与身份

系统通过请求头 `X-User-ID` 标识当前用户。迁移初始化了以下用户：

| ID | 角色 | 名称 |
| --- | --- | --- |
| 1 | customer | Customer Alice |
| 2 | agent | Agent Bob |
| 3 | supervisor | Supervisor Carol |
| 4 | agent | Agent Dana |

## API

| 方法 | 路径 | 角色 | 说明 |
| --- | --- | --- | --- |
| POST | `/api/v1/tickets` | customer | 提交工单 |
| GET | `/api/v1/tickets` | 所有角色 | 查看工单列表，支持 `status`、`priority`、`assignee_id` 筛选；客户只能看自己的工单 |
| GET | `/api/v1/tickets/:id` | 所有角色 | 查看工单详情 |
| PATCH | `/api/v1/tickets/:id/assign` | supervisor | 分配工单给客服 |
| PATCH | `/api/v1/tickets/:id/claim` | agent | 客服领取工单 |
| PATCH | `/api/v1/tickets/:id/status` | agent, supervisor | 更新状态，可同时填写结果和备注 |
| PATCH | `/api/v1/tickets/:id/result` | agent, supervisor | 填写处理结果和备注 |
| PATCH | `/api/v1/tickets/:id/priority` | supervisor | 设置优先级 |
| GET | `/api/v1/tickets/:id/history` | 所有角色 | 查看状态变更历史 |
| GET | `/api/v1/statistics` | supervisor | 查看工单统计 |

## 状态

状态机定义在 `internal/service`：

```text
open -> in_progress | pending | resolved | closed
in_progress -> pending | resolved | closed
pending -> in_progress | resolved | closed
resolved -> closed | open
closed -> open
```

每次状态变化都会写入 `ticket_status_history`。工单分配和领取使用事务行锁，避免并发重复分配。

## SLA

列表返回的每个工单会包含 `sla_status` 和 `sla_breached`。不同优先级的 SLA 时限如下：

| 优先级 | SLA |
| --- | --- |
| urgent | 4 小时 |
| high | 8 小时 |
| normal | 24 小时 |
| low | 48 小时 |
