
# type VarQueue
goroutine間通信用の可変長キューのモジュールです。

## import
```go
import "github.com/l4go/vqueue"
```
vendoringして使うことを推奨します。

## サンプルコード

[example](../examples/ex_vqueue/ex_vqueue.go)

## メソッド概略
### func New(free func(interface{})) *VarQueue
キューを生成します。引数'free'には、Close()で処分されるデータの処分方法を指定します。
処分の処理が必要ないときはnilを指定します。

### func (vq *VarQueue) Close()
キューを終了します。キューに残っているデータはPop()で取り出されるのを待ちます。

### func (vq *VarQueue) Cancel()
キューを終了します。キューに残っているデータは処分されて、空にされます。

### func (vq *VarQueue) Shrink()
現在の保存データサイズまでキューのメモリサイズを圧縮します。

### func (vq *VarQueue) Pop() (interface{}, bool)
キューからデータを取り出します。キューが空の時は、ブロックします。
データが取り出せた場合は、取り出せたデータと、trueが返されます。
Close()された場合は、nilとfalseが返されます。

### func (vq *VarQueue) PopWithCancel(cc task.Canceller) (interface{}, bool)
キューからデータを取り出します。
ccがキャンセルされた場合は、処理を中断します。
データが取り出せた場合は、取り出せたデータと、trueが返されます。
Close()された場合もしくは、キューが空の場合は、nilとfalseが返されます。

### func (vq *VarQueue) PopNonblock() (interface{}, bool)
キューからデータを取り出します。
データが取り出せた場合は、取り出せたデータと、trueが返されます。
Close()された場合もしくは、キューが空の場合は、nilとfalseが返されます。

### func (vq *VarQueue) PopOrTimeout(d time.Duration) (interface{}, bool, bool)
キューからデータを取り出します。キューが空の時は、タイムアウト時間までブロックします。
２つのbool値はそれぞれ、値が取得できたこと、タイムアウトしたことを意味します。

タイムアウトした時は、nil, false, trueが返されます。
データが取り出せた場合は、取り出せたデータ、true, falseが返されます。
すでにClose()された場合は、nil, false, falseが返されます。

## func (vq *VarQueue) Push(v interface{})
キューにデータを入れます。ブロックしません。
