package cron

import (
	"github.com/robfig/cron/v3"

	"github.com/nostack-cn/svr-admin/task"
)

// Scheduler 全局定时任务调度器
var Scheduler *cron.Cron

// Setup 初始化并注册所有定时任务
func Setup() {
	Scheduler = cron.New(cron.WithSeconds()) // 支持秒级 cron 表达式

	// 注册定时任务
	registerTasks()

	Scheduler.Start()
}

func registerTasks() {
	// 每 60 秒执行一次示例任务
	_, _ = Scheduler.AddFunc("0 */1 * * * *", task.ExampleTask)
}
