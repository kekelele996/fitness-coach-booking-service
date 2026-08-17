# BUG_REPRO

## Bug 是什么
retrying 状态在转换表、重试写回、活跃过滤、调度查询四处未同步：转换表缺 retrying->confirmed、RetryBooking 写回 failed、FilterActive 漏 retrying、worker 只找 confirmed。

## 如何触发
`cd <env> && go test -count=20 ./internal/...`

## 错误信息
- model.TestCanTransition / TestActiveStatusesIncludesRetrying 失败。
- service.TestRetryLifecycle 失败（重试后仍 failed）。
- util.TestFilterActive 失败。
- worker.TestTickRetriesFailed 失败（捞不到 failed）。
