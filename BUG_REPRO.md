Bug 是什么

考勤规则配置的校验、启用规则选择、规则更新和创建接口处理不一致，非法上下班时间能通过，启用的弹性规则没有被应用。

如何触发

在项目根目录运行：

```bash
go test ./...
```

错误信息

```text
TestRuleValidate/下班早于上班: 应校验失败
TestCreateRuleValidation: 下班早于上班应被拒绝
TestActiveRuleReturnsEnabled: 应返回启用的规则B，实际 08:30-17:30
TestAttendanceUsesCustomRule: 10:20 应为 normal，实际 late
```
