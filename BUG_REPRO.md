Bug 是什么

考勤相关流程在迟到阈值、日期范围筛选、补卡唯一性和签退接口上出现联动错误，导致正常阈值内打卡被判迟到，日期过滤漏数据，重复日期补卡未被拒绝，签退接口返回了签到状态。

如何触发

在项目根目录运行：

```bash
go test ./...
```

错误信息

```text
TestCheckOutAPI: 17:00 签退应为 early_leave，实际 late
TestAttendanceFilterMatch: 日期范围内应匹配
TestAttendanceUsesDefaultRule: 默认规则下 09:10 应为 normal，实际 late
TestAttendanceUpdateUniqueDate: 更新为同日考勤应返回 ErrConflict，实际: <nil>
```
