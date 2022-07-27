
# 实验一

* 你应该把你的实现放在mr/coordinator.go、mr/worker.go和mr/rpc.go中
* demo:
  src/main/mrsequential.go 参考
* 协调器 main/mrcoordinator.go 不要改变
* 工作器的 "主 "进程在 main/mrworker.go 不要改

* 测试用例
  - word-count in： mrapps/wc.go
  - and a text indexer in： mrapps/indexer.go


* 运行
```bash

cd src/main
go build -race -buildmode=plugin ../mrapps/wc.go
rm mr-out*
go run -race mrsequential.go wc.so pg*.txt
more mr-out-0
```