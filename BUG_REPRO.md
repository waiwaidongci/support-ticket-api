# Bug 复现说明

## Bug 是什么
不允许的状态流转错误没有包住 `ErrInvalidTransition`，上层无法正确识别这个错误。

## 如何触发
把一张 `in_progress` 工单改回 `open`。

## 错误信息
运行 `go test ./...` 会失败，期望 `errors.Is(ErrInvalidTransition)`，实际匹配不到。
