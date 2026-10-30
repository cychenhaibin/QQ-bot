# QQ Bot Service

独立的官方 QQ 开放平台机器人服务。它通过 QQ Gateway WebSocket 接收QQ群消息，通过 QQ OpenAPI 回复消息，后续再通过受保护的 new-api 内部接口完成用户、签到和 AI 能力调用。

## 本阶段能力

- 官方 QQ Bot AppID/AppSecret 鉴权
- 获取 Gateway 地址和 Access Token
- Gateway Hello、Identify、心跳和群消息事件接收
- 识别 `GROUP_AT_MESSAGE_CREATE` / `GROUP_MESSAGE_CREATE`
- 用户发送 `菜单`、`/菜单` 或 `帮助` 时回复功能菜单
- 调用 new-api 的 HMAC 内部请求客户端基础能力
- `[绑定账号]` 邮箱验证码交互流程和换绑确认流程
- `[查询余额]` 返回已绑定用户名和当前余额
- `[每日签到]` 调用 new-api 签到并回复固定成功提示
- `[充值中心]` 返回充值链接并提示区分付费站和公益站
- `[模型渠道]` 只读返回可用分组和模型名称，不暴露渠道 Key、Base URL 等敏感配置
- `[AI 对话]` 按群成员隔离会话，依次选择分组和模型，自动发送首句 `hi`，支持“停”结束会话
- 每个功能开始前选择 `FluxA` 或 `Charity`，两个站点连接独立的 new-api 实例；多步骤流程会沿用本次选择，下一个功能重新选择

绑定流程调用 new-api 的受保护接口：

```text
POST /api/bot/v1/bindings/start
POST /api/bot/v1/bindings/verify
POST /api/bot/v1/account/lookup
POST /api/bot/v1/checkin
POST /api/bot/v1/channels
POST /api/bot/v1/ai/groups
POST /api/bot/v1/ai/access
```

充值中心不调用 new-api，直接由独立 Bot 服务返回当前站点配置的充值链接：

```text
https://catfk.com/shop/FluxA
```

AI 对话拿到分组对应的 API Key 后，通过 new-api 的兼容接口发送请求：

```text
POST /v1/chat/completions
```

验证码和邮箱归属校验、用户余额读取、签到扣增额度均由 new-api 完成。Bot 服务只在内存中保存短期 flow token，不记录验证码和服务密钥。Bot 服务和 new-api 通过 HMAC 内部接口通信，两个项目保持独立目录和 Go 模块。

## 配置

```bash
export QQ_BOT_APP_ID="your-app-id"
export QQ_BOT_APP_SECRET="your-app-secret"

# 可选；默认使用官方地址
export QQ_BOT_API_BASE_URL="https://api.bot.qq.com"
export QQ_BOT_TOKEN_ENDPOINT="https://bots.qq.com/app/getAppAccessToken"
export QQ_BOT_GATEWAY_URL=""

# FluxA new-api
export FLUXA_NEW_API_BASE_URL="https://your-fluxa-new-api.example.com"
export FLUXA_NEW_API_BOT_ID="qq-bot-fluxa"
export FLUXA_NEW_API_SERVICE_SECRET="your-fluxa-internal-service-secret"
export FLUXA_RECHARGE_URL="https://catfk.com/shop/FluxA"

# Charity new-api
export CHARITY_NEW_API_BASE_URL="https://your-charity-new-api.example.com"
export CHARITY_NEW_API_BOT_ID="qq-bot-charity"
export CHARITY_NEW_API_SERVICE_SECRET="your-charity-internal-service-secret"
export CHARITY_RECHARGE_URL="https://your-charity-recharge.example.com"
```

## 启动

```bash
go mod tidy
go run ./cmd/bot
```

生产环境可以使用 Docker：

```bash
docker build -t qq-bot-service .
docker run --env-file .env qq-bot-service
```

## 验证

```bash
go test ./...
go vet ./...
```
