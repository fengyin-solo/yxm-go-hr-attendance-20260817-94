Bug 是什么

加班单在时长校验、审批状态、删除持久化和驳回接口之间出现组合错误，零时长被放行，审批通过后统计与月报都看不到有效加班。

如何触发

在项目根目录运行：

```bash
go test ./...
```

错误信息

```text
TestOvertimeValidate/零时长: 应校验失败
TestOvertimeLifecycle: 审批后应为 approved，实际 rejected
TestEmployeeMonthlyReport: 加班时长应为 2，实际 0
TestOvertimeStats: 加班统计错误: &{Count:0 TotalHours:0}
```
