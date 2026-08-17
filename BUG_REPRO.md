Bug 是什么

请假审批链路的状态流转、审批接口、删除持久化和批量审批结果不一致，审批通过后会落成错误状态，已审批单据的保护也失效。

如何触发

在项目根目录运行：

```bash
go test ./...
```

错误信息

```text
TestLeaveApproveAPI: 审批后应为 approved，实际 rejected
TestLeaveTransitions: 流转 approved -> rejected 期望 false，实际 true
TestLeaveLifecycle: 审批后应为 approved，实际 rejected
TestLeaveCRUD: 删除后应返回 ErrNotFound
```
