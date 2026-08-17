# Bug 复现说明

## Bug 是什么
统计结果里的三个映射字段在空数据时保持为 nil，JSON 序列化后会变成 `null`，而不是 `{}`。

## 如何触发
以 supervisor 身份请求统计接口，并让 `by_status`、`by_priority`、`by_assignee` 都没有条目。

## 错误信息
运行下面的测试会失败，提示期望非 nil map。

```bash
cd internal && go test ./service -run TestStatisticsAlwaysReturnsJSONMaps -count=1
```
