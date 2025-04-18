# Airtable-like API 文档

## 基础资源接口

### Base 管理

| 方法   | 路径                | 描述          | 状态码       |
|--------|---------------------|---------------|-------------|
| POST   | /api/v1/bases       | 创建新Base    | 201 Created |
| GET    | /api/v1/bases       | 获取所有Base  | 200 OK      |
| GET    | /api/v1/bases/{baseId} | 获取单个Base | 200 OK      |
| PUT    | /api/v1/bases/{baseId} | 更新Base     | 200 OK      |
| DELETE | /api/v1/bases/{baseId} | 删除Base     | 204 No Content |

**请求示例**：
```json
POST /api/v1/bases
{
  "name": "项目空间",
  "description": "团队协作空间"
}
```

## 表格管理接口

### Table 操作

| 方法   | 路径                                | 描述            |
|--------|-------------------------------------|-----------------|
| POST   | /api/v1/bases/{baseId}/tables      | 创建新表格      |
| GET    | /api/v1/bases/{baseId}/tables      | 获取所有表格    |
| GET    | /api/v1/bases/{baseId}/tables/{tableId} | 获取单个表格    |
| PUT    | /api/v1/bases/{baseId}/tables/{tableId} | 更新表格信息    |
| DELETE | /api/v1/bases/{baseId}/tables/{tableId} | 删除表格        |

**响应示例**：
```json
{
  "id": "c3d4e5f6-7890-1234-5678-9abcdef01234",
  "name": "用户表",
  "fields": []
}
```

## 字段管理接口

### Field 操作

| 方法   | 路径                                        | 描述            |
|--------|---------------------------------------------|-----------------|
| POST   | /api/v1/bases/{baseId}/tables/{tableId}/fields | 创建新字段      |
| GET    | /api/v1/bases/{baseId}/tables/{tableId}/fields | 获取所有字段    |
| PUT    | /api/v1/bases/{baseId}/tables/{tableId}/fields/{fieldId} | 更新字段        |
| DELETE | /api/v1/bases/{baseId}/tables/{tableId}/fields/{fieldId} | 删除字段        |

**字段类型支持**：
- text
- number
- boolean
- date

## 错误代码

| 状态码 | 描述                  |
|--------|-----------------------|
| 400    | 无效请求参数          |
| 401    | 未授权访问            |
| 404    | 资源不存在            |
| 409    | 字段/表名称冲突       |
| 422    | 参数验证失败          |
| 500    | 服务器内部错误        |

## 记录管理接口

### Record 操作

| 方法   | 路径                                        | 描述                |
|--------|---------------------------------------------|---------------------|
| POST   | /api/v1/bases/{baseId}/tables/{tableId}/records | 创建新记录          |
| GET    | /api/v1/bases/{baseId}/tables/{tableId}/records | 查询记录（支持过滤）|
| GET    | /api/v1/bases/{baseId}/tables/{tableId}/records/{recordId} | 获取单个记录        |
| PUT    | /api/v1/bases/{baseId}/tables/{tableId}/records/{recordId} | 更新记录            |
| DELETE | /api/v1/bases/{baseId}/tables/{tableId}/records/{recordId} | 删除记录            |

**查询参数**：
- filter: 过滤条件（JSON格式）
- sort: 排序字段
- maxRecords: 最大返回数量

## WebSocket接口

| 路径 | 描述                          |
|------|-------------------------------|
| /ws  | 实时数据变更通知              |

## 健康检查

| 路径    | 描述         |
|---------|--------------|
| /health | 服务状态检查 |

## 测试说明

测试用例需要验证：
1. 资源创建时的参数校验
2. 级联删除功能
3. 字段类型约束
4. 唯一性约束检查
5. 记录查询过滤逻辑
6. WebSocket连接稳定性

```go
// WebSocket测试示例
func TestWebSocketUpdates(t *testing.T) {
  // 测试数据变更通知机制
}

// 记录查询测试
func TestRecordFilter(t *testing.T) {
  // 测试过滤条件解析逻辑
}
```
```go
// 示例测试用例
func TestCreateTable(t *testing.T) {
  // 测试表格创建逻辑
}
```