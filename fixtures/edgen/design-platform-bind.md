# Edgen 服务端设计 — 平台绑定模块

## 服务

- 主 API：`api.edgen.tech`（PROD）、`api.beta.ospprotocol.xyz`（BETA）
- 路径前缀：`/v2/platform/`

## 接口清单（本迭代 MR-EDGEN-DEMO）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/v2/platform/bind/{platform}` | 查询指定平台（如 TWITTER）绑定状态 |
| GET | `/v2/platform/bindings` | 分页查询用户全部平台绑定 |
| DELETE | `/v2/platform/bind/{platform}` | 解绑（本期可不测） |

## GET /v2/platform/bind/{platform}

- **鉴权**：`Authorization: Bearer <token>`
- **路径参数**：`platform` 枚举 `TWITTER` | `DISCORD` | …
- **Query**：`reverse` 可选 boolean，默认 false

### 响应 200

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "platform": "TWITTER",
    "bound": true,
    "externalId": "123456",
    "boundAt": "2026-01-15T08:00:00Z"
  }
}
```

### 错误码

- `401` 未授权
- `404` 不支持的平台类型

## GET /v2/platform/bindings

- Query：`page`, `pageSize`（默认 20，最大 50）
- 响应 `data.rows[]` 含 `platform`, `bound`, `boundAt`
