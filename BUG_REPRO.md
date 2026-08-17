# Bug 复现说明

## Bug 是什么
创建工单的入参校验错误没有包住 `ErrInvalidInput`，调用方用 `errors.Is` 判断不到这个错误。

## 如何触发
调用创建工单接口，把 `title` 留空，或传入不支持的 `type` / `priority`。

## 错误信息
运行下面的测试会失败，报错 `want errors.Is(ErrInvalidInput)`。

```bash
go test ./internal/service -run TestCreateTicketValidationWrapsInvalidInput -count=1
```
