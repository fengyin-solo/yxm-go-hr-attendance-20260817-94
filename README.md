# HR Attendance 人事考勤服务

一个纯 Go 标准库实现的人事考勤后端，覆盖部门、员工、考勤、请假、加班与考勤规则全流程，内置订单级状态机、考勤状态自动判定、多维度报表、鉴权与限流中间件，支持批量导入与公司级数据导出。

## 运行

```bash
go run ./cmd/server
```

环境变量：

| 变量 | 默认值 | 说明 |
|------|-------|------|
| `PORT` | 8080 | 监听端口 |
| `ADDR` | `:8080` | 完整监听地址（优先于 PORT） |
| `MAX_PAGE_SIZE` | 100 | 单页最大条数 |
| `ADMIN_TOKEN` | 空 | 管理令牌，非空时写操作需鉴权 |
| `RATE_LIMIT` | 100 | 每窗口最大请求数（按客户端 IP） |
| `RATE_WINDOW_SEC` | 60 | 限流窗口秒数 |
| `LATE_THRESHOLD_MIN` | 15 | 默认迟到阈值（分钟） |
| `DEFAULT_WORK_START` | 09:00 | 默认上班时间 |
| `DEFAULT_WORK_END` | 18:00 | 默认下班时间 |
| `LOG_LEVEL` | info | 日志级别：debug/info/warn/error |

## 鉴权

设置 `ADMIN_TOKEN` 后，所有写操作（POST/PUT/DELETE）需携带请求头：

```
Authorization: Bearer <ADMIN_TOKEN>
```

读操作（GET）不受限制。未设置 `ADMIN_TOKEN` 时全部放行。

## API 一览

统一响应结构：`{"code":0,"message":"ok","data":...}`；错误时 `code` 非 0 且 `message` 说明原因。

### 部门 Department

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/departments` | 创建部门 |
| GET | `/api/departments` | 分页列出（parent_id/status/keyword） |
| GET | `/api/departments/{id}` | 获取部门详情 |
| PUT | `/api/departments/{id}` | 更新部门 |
| DELETE | `/api/departments/{id}` | 删除部门（有子部门或员工时拒绝） |
| GET | `/api/departments/{id}/employee-count` | 统计在职员工数 |

### 员工 Employee

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/employees` | 创建员工（工号唯一） |
| GET | `/api/employees` | 分页列出（department_id/status/keyword） |
| GET | `/api/employees/{id}` | 获取员工详情 |
| PUT | `/api/employees/{id}` | 更新员工（转部门/离职） |
| DELETE | `/api/employees/{id}` | 删除员工（有考勤记录时拒绝） |
| POST | `/api/employees/batch` | 批量导入员工 |

### 考勤 Attendance

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/attendances` | 创建考勤记录（自动判定状态） |
| GET | `/api/attendances` | 分页列出（employee_id/status/from/to） |
| GET | `/api/attendances/{id}` | 获取考勤详情 |
| PUT | `/api/attendances/{id}` | 补卡（重算状态） |
| DELETE | `/api/attendances/{id}` | 删除考勤 |
| POST | `/api/attendances/check-in` | 签到 |
| POST | `/api/attendances/check-out` | 签退 |
| POST | `/api/attendances/batch` | 批量导入考勤 |

考勤状态由规则自动判定：`normal`（正常）、`late`（迟到）、`early_leave`（早退）、`absent`（缺勤）。

### 请假 Leave

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/leaves` | 创建请假单（类型 annual/sick/personal） |
| GET | `/api/leaves` | 分页列出（employee_id/type/status/from/to） |
| GET | `/api/leaves/{id}` | 获取请假单 |
| DELETE | `/api/leaves/{id}` | 删除（已审批通过的不可删除） |
| POST | `/api/leaves/{id}/approve` | 审批通过 |
| POST | `/api/leaves/{id}/reject` | 驳回 |
| POST | `/api/leaves/batch-approve` | 批量审批通过 |

状态机：`pending → approved / rejected`。

### 加班 Overtime

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/overtimes` | 创建加班单 |
| GET | `/api/overtimes` | 分页列出（employee_id/status/from/to） |
| GET | `/api/overtimes/{id}` | 获取加班单 |
| DELETE | `/api/overtimes/{id}` | 删除（已审批通过的不可删除） |
| POST | `/api/overtimes/{id}/approve` | 审批通过 |
| POST | `/api/overtimes/{id}/reject` | 驳回 |
| POST | `/api/overtimes/batch-approve` | 批量审批通过 |

### 考勤规则 Rule

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/rules` | 创建规则 |
| GET | `/api/rules` | 列出全部规则 |
| GET | `/api/rules/{id}` | 获取规则详情 |
| PUT | `/api/rules/{id}` | 更新规则 |
| DELETE | `/api/rules/{id}` | 删除规则 |

### 报表 Report

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/reports/attendance-status` | 考勤状态分布 |
| GET | `/api/reports/leave-types` | 请假类型统计（已审批天数） |
| GET | `/api/reports/overtime` | 加班统计 |
| GET | `/api/reports/company` | 公司级汇总报告 |
| GET | `/api/reports/top-late?month=&limit=` | 月度迟到排行 |
| GET | `/api/reports/perfect-attendance?month=` | 月度全勤员工 |
| GET | `/api/employees/{id}/monthly-report?month=` | 员工月度考勤汇总 |
| GET | `/api/departments/{id}/monthly-report?month=` | 部门月度汇总 |
