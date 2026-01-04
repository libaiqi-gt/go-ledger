# 接口文档（go-ledger）

本项目使用 Gin 构建 REST API，数据库使用 MySQL，ORM 使用 GORM。当前所有接口统一前缀为 `/v1`。带鉴权的接口需在请求头携带 `Authorization: Bearer <token>`。

## 基本信息
- 基础路径：`/v1`
- 内容类型：`application/json`
- 身份认证：部分接口需要 JWT，登录后获取 `token`
- 错误返回：统一使用 `{"error": "<错误描述>"}` 或具体业务字段

## 路由概览
- 公开接口
  - `POST /v1/register` 用户注册
  - `POST /v1/login` 用户登录，返回 JWT token
- 需鉴权接口（需 `Authorization: Bearer <token>`）
  - `POST /v1/entries` 创建账单
  - `GET /v1/entries` 分页查询当前用户账单（支持筛选）
  - `DELETE /v1/entries/:id` 删除账单

路由定义参考：[router.go](file:///d:/GO/go-ledger/routers/router.go)

---

## 1. 用户注册
- 方法与路径：`POST /v1/register`
- 说明：创建新用户（用户名唯一）
- 请求体：

```json
{
  "username": "alice",
  "password": "P@ssw0rd"
}
```

- 成功响应：

```json
{
  "message": "注册成功",
  "data": "alice"
}
```

- 失败示例：

```json
{
  "error": "注册失败，用户名可能已存在"
}
```

实现参考：[auth.go:Register](file:///d:/GO/go-ledger/controllers/auth.go#L18-L56)

---

## 2. 用户登录
- 方法与路径：`POST /v1/login`
- 说明：用户名密码校验，返回 JWT token
- 请求体：

```json
{
  "username": "alice",
  "password": "P@ssw0rd"
}
```

- 成功响应：

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

- 失败示例：

```json
{
  "error": "用户名或密码错误"
}
```

实现参考：[auth.go:Login](file:///d:/GO/go-ledger/controllers/auth.go#L58-L83)  
Token 生成参考：[token.go:GenerateToken](file:///d:/GO/go-ledger/utils/token.go#L12-L27)  
JWT 配置参考（密钥）：[config.yaml](file:///d:/GO/go-ledger/config/config.yaml#L7-L8)

---

## 3. 新增账单
- 方法与路径：`POST /v1/entries`
- 说明：为当前登录用户新增一条账单记录
- 鉴权：需要
- 请求头：
  - `Authorization: Bearer <token>`
- 请求体（建议字段与类型）：

```json
{
  "type": 1,
  "amount": 199.99,
  "category": "餐饮",
  "date": "2025-01-01",
  "remark": "朋友聚餐"
}
```

- 字段说明（与模型对应）：
  - `type`：整数，`1` 收入，`2` 支出
  - `amount`：数字，建议两位小数
  - `category`：字符串
  - `date`：日期字符串，建议 `YYYY-MM-DD`
  - `remark`：字符串（可选）

- 成功响应：

```json
{
  "data": {
    "id": 1,
    "user_id": 10,
    "type": 1,
    "amount": 199.99,
    "category": "餐饮",
    "date": "2025-01-01",
    "remark": "朋友聚餐",
    "created_at": "2025-01-01T12:00:00Z",
    "updated_at": "2025-01-01T12:00:00Z"
  }
}
```

实现参考：[entry.go:CreateEntry](file:///d:/GO/go-ledger/controllers/entry.go#L9-L27)  
模型参考：[models/entry.go](file:///d:/GO/go-ledger/models/entry.go)

---

## 4. 查询账单列表
- 方法与路径：`GET /v1/entries`
- 说明：查询当前登录用户的账单，支持筛选与分页，按日期倒序返回
- 鉴权：需要
- 请求头：
  - `Authorization: Bearer <token>`
- 查询参数（全部可选）：
  - `type`：整数，`1` 收入，`2` 支出
  - `category`：字符串，类别（精确匹配）
  - `start_date`：字符串，开始日期（`YYYY-MM-DD`）
  - `end_date`：字符串，结束日期（`YYYY-MM-DD`）
  - `page`：当前页（默认 `1`）
  - `page_size`：每页数量（默认 `10` 或项目约定值）
- 响应示例：

```json
{
  "data": [
    {
      "id": 1,
      "user_id": 10,
      "type": 2,
      "amount": 88.00,
      "category": "交通",
      "date": "2025-01-02",
      "remark": "地铁"
    }
  ],
  "meta": {
    "current_page": 1,
    "page_size": 10,
    "total": 23,
    "total_pages": 3
  }
}
```

实现参考：[entry.go:FindEntries](file:///d:/GO/go-ledger/controllers/entry.go#L31-L84)  
分页工具（在代码中引用）：`utils.Paginate`、`utils.GetPageParams`（用于计算分页与 `meta` 返回）

---

## 5. 删除账单
- 方法与路径：`DELETE /v1/entries/:id`
- 说明：删除指定 ID 的账单
- 鉴权：需要
- 请求头：
  - `Authorization: Bearer <token>`
- 路径参数：
  - `id`：账单 ID（整数）
- 成功响应：

```json
{
  "data": "删除成功"
}
```

实现参考：[entry.go:DeleteEntry](file:///d:/GO/go-ledger/controllers/entry.go#L39-L44)

---

## 中间件与鉴权
- JWT 鉴权中间件：从请求头 `Authorization` 提取 `Bearer <token>`，解析校验后写入 `userID` 到上下文
- 参考实现：[middlewares/auth.go](file:///d:/GO/go-ledger/middlewares/auth.go)

---

## 配置
- 配置文件路径：`config/config.yaml`
- 数据库连接字段：
  - `database.host`、`database.port`、`database.user`、`database.password`、`database.dbname`
- JWT：
  - `jwt.secret`：用于签名 Token
- 读取配置参考：[config/database.go:InitConfig](file:///d:/GO/go-ledger/config/database.go#L13-L22)
