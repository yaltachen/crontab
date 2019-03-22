# 两种获取channel方式
```go
// 使用for range方式，当外部channel被关闭的时候，for会退出
// range可以识别出channel是否被关闭
for x := range ch{
    ...
}
```
```go
// 使用<-ch，channel被关闭时，会不断读出初始值
select {
    case x = <-ch:
    ...
}
```