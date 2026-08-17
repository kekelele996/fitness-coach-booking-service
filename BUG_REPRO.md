# BUG_REPRO

## Bug 是什么
存储层读路径去掉了深拷贝，GetBooking / ListBookings / GetPackage 直接返回内部引用；service 的 ActiveCount 在 goroutine 内 Add WaitGroup 并并发读这些引用；worker 重试循环去掉了 Attempts < MaxAttempts 上限。

## 如何触发
`cd <env> && go test ./... -race -count=20`

## 错误信息
- store 单测报“返回内部引用被改坏”（TestGetBookingReturnsCopy / TestGetPackageReturnsCopy）。
- `-race` 在 service.TestActiveCount 报 DATA RACE（WaitGroup 计数错配、并发读写）。
- worker.TestTickRespectsRetryLimit 失败：到达重试上限仍被重试。
