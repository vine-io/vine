# gscheduler

## 简介

gscheduler 是`golang`实现的一个简单的任务调度器。 实现功能：

- 任务的添加
- 任务的删除
- 任务的修改
- 任务的查询

## 实例

```go
package main

import (
	"fmt"
	"time"

	"github.com/vine-io/gscheduler"
)

func main() {
	scheduler := gscheduler.NewScheduler()
	scheduler.Start()

	a := 1

	job1, _ := gscheduler.JobBuilder().Name("job1").Duration(time.Second).Fn(func() {
		fmt.Printf("[%s] a = %d\n", time.Now(), a)
		a++
	}).Out()

	job2, _ := gscheduler.JobBuilder().Name("job2").Duration(time.Second).Fn(func() {
		fmt.Println("job2", time.Now())
	}).Out()

	// 添加任务
	scheduler.AddJob(job1)
	scheduler.AddJob(job2)
	
	// 删除任务
	scheduler.RemoveJob(job1)
    
	// 修改任务
    scheduler.UpdateJob(job2)
	
	// 查询任务
	scheduler.GetJob(job2.ID())

	time.Sleep(time.Second * 10)
}
```

## 构建任务
```go
func main() {
	// 构建一个符合正则表达式的任务
	gscheduler.JobBuilder().Name("cron-job").Spec("*/10 * * * * * *").Out()

	// 构建一次性延时任务
	gscheduler.JobBuilder().Name("delay-job").Delay(time.Now().Add(time.Hour * 3)).Out()

	// 构建间隔执行的任务
	gscheduler.JobBuilder().Name("duration-job").Duration(time.Second * 10).Out()
	
	// 构建多次执行的任务
	gscheduler.JobBuilder().Name("three-times-job").Duration(time.Second*5).Times(3).Out()
}
```
