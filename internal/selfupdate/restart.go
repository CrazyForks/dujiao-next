package selfupdate

import (
	"os"
	"syscall"
	"time"
)

// restartDelay 触发退出前的等待时间，保证 HTTP 响应已经写回客户端。
// 前端需要先拿到「正在重启」的结果才能进入轮询等待状态。
const restartDelay = 500 * time.Millisecond

// Restart 让当前进程退出，由 systemd 按 Restart= 策略拉起新二进制。
//
// 走 SIGTERM 而不是 os.Exit：cmd/server 用 signal.NotifyContext 监听 SIGTERM，
// 收到后会依次 Stop 各个 Service —— HTTP server 优雅关闭、asynq worker 等待在途任务。
// 直接 os.Exit 会让正在处理的订单请求和后台任务被硬切断。
//
// 调用前必须确认 Detect().CanRestart 为 true。没有守护进程时退出即停服，
// 这种情况下应由前端提示用户手动重启，而不是调用本函数。
func Restart() {
	go func() {
		time.Sleep(restartDelay)
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			// 理论上对自身 PID 不会失败；真失败了就直接退出，
			// 让 systemd 拉起新进程，代价是跳过优雅关闭。
			os.Exit(0)
			return
		}
		if err := p.Signal(syscall.SIGTERM); err != nil {
			os.Exit(0)
		}
	}()
}
