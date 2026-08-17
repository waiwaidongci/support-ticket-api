# Bug 复现说明

## Bug 是什么
列表构建时先按原长度预分配切片，又继续 append，导致前面多出与原始条数相同的零值元素。

## 如何触发
准备两条工单，请求工单列表接口。

## 错误信息
运行下面的测试会失败，实际返回 4 条，前两条是空对象。

```bash
go test -run TestListTicketsBuildsExactItems -count=1 ./...
```
