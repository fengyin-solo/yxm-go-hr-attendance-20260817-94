Bug 是什么

员工资料维护和部门统计链路出现多处不一致：工号关键词搜不到员工，在职员工数统计为 0，员工更新后的资料也没有稳定反映到查询结果。

如何触发

在项目根目录运行：

```bash
go test ./...
```

错误信息

```text
TestDepartmentEmployeeCountAPI: 员工数应为 2，实际 0
TestEmployeeFilterMatch: 按工号关键词应匹配
TestDepartmentEmployeeCount: 在职员工数应为 2，实际 0
TestEmployeeKeywordFilter: 按工号关键词筛选结果错误: total=0
```
