# fitness-coach-booking-service

一个用 Go 写的健身私教课时预约与扣减服务，演示 `handler → service → repository → store → model` 分层、预约状态机与后台调度 worker。

## 功能

- 会员与私教教练管理，教练按技能标签派课
- 会员购买课时包，预约私教课自动扣减课时
- 预约状态流转：pending → confirmed → completed / cancelled / expired → failed → retrying
- 取消预约自动退还课时
- 后台调度器过期未确认预约、重试失败预约
- 内存存储，读写锁保护，支持按状态查询与统计

## 目录结构

```
cmd/server/           程序入口
internal/config/      环境变量配置
internal/model/       模型定义与状态机
internal/store/       内存存储
internal/repository/  数据访问层
internal/service/     业务逻辑（建单/扣课时/确认/取消/查询）
internal/worker/      后台调度 worker
internal/handler/     HTTP 接口
internal/util/        过滤/排序等工具函数
```

## 运行与测试

```bash
go build ./...
go test ./...
go run ./cmd/server
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `FITNESS_WORKERS` | 调度 worker 数量 | `4` |
| `FITNESS_RETRY_LIMIT` | 失败预约重试上限 | `3` |
| `FITNESS_EXPIRE_MINUTES` | 预约过期分钟数 | `30` |
| `FITNESS_POLL_INTERVAL_MS` | 调度轮询间隔（毫秒） | `500` |
| `FITNESS_SKILL_ROUTES` | 训练项目到技能标签路由 | 内置默认 |

## 技术栈

- Go 1.22
