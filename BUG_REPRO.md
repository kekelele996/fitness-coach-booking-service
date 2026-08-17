# BUG_REPRO

## Bug 是什么
repository 层把哨兵错误用 %v 包装丢了 %w 链，service 层又用错误哨兵判断，handler 把 not found 映射成 500，worker 遇 not found 不跳过。

## 如何触发
`cd <env> && go test ./... -count=20`

## 错误信息
- repository.TestFindBookingWrapsNotFound / TestFindCoachWrapsNotFound 失败（errors.Is 失效）。
- handler.TestGetMissingBookingReturns404 失败（500 而非 404）。
- service.TestFindMissingWrapsNotFound 失败。
