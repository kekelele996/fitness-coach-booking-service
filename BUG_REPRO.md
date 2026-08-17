# BUG_REPRO

## Bug 是什么
FilterBookingsByStatus / FilterActive 用 bookings[:0] 原地复用底层数组，Stats 在同一份快照上连续过滤读到被污染切片。

## 如何触发
`cd <env> && go test ./internal/service/ -run TestStats`

## 错误信息
- service.TestStats 失败：confirmed 数被污染成 0（应 1）。
- util.TestFilterByStatusNoAliasing 失败（入参被改写）。
