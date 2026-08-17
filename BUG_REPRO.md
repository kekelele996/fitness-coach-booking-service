# BUG_REPRO

## Bug 是什么
自定义技能路由解析为空时返回 nil map，构造函数去掉了 nil 兜底，SkillFor 向 nil map 写入触发 panic。

## 如何触发
`cd <env> && go test -run TestSkillRouteFallbackDispatch ./...`

## 错误信息
- service.TestSkillRouteFallbackDispatch panic: assignment to entry in nil map。
- config.TestSkillRouteFallback 失败（routes 为空）。
