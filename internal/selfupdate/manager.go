package selfupdate

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/dujiao-next/internal/version"
)

// Stage 升级过程中的阶段，用于前端展示进度条文案
type Stage string

const (
	StageIdle        Stage = "idle"
	StageDownloading Stage = "downloading"
	StageVerifying   Stage = "verifying"
	StageExtracting  Stage = "extracting"
	StageSwapping    Stage = "swapping"
	StageDone        Stage = "done"
)

// Status 升级任务的整体状态
type Status string

const (
	StatusIdle      Status = "idle"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

// ErrUpdateInProgress 已有升级任务在执行
var ErrUpdateInProgress = errors.New("update already in progress")

// State 升级任务状态快照，直接序列化给前端轮询
type State struct {
	Status  Status `json:"status"`
	Stage   Stage  `json:"stage"`
	Percent int    `json:"percent"`
	// TargetVersion 本次升级的目标版本号
	TargetVersion string `json:"target_version,omitempty"`
	// Error 失败原因（英文技术细节），前端按 status 决定是否展示
	Error string `json:"error,omitempty"`
	// RestartRequired 二进制已替换但进程仍是旧版本，需要重启才生效
	RestartRequired bool       `json:"restart_required"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
}

// Manager 串行化升级任务并维护可轮询的状态。
// 升级过程跨越单次 HTTP 请求（下载可能几分钟），因此任务在后台 goroutine 执行，
// 前端通过 Snapshot 轮询进度。
type Manager struct {
	mu      sync.Mutex
	state   State
	running bool
	updater *Updater
	// now 便于测试注入时间
	now func() time.Time
}

// NewManager 创建升级任务管理器
func NewManager() *Manager {
	return &Manager{
		state:   State{Status: StatusIdle, Stage: StageIdle},
		updater: NewUpdater(),
		now:     time.Now,
	}
}

// Snapshot 返回当前状态副本
func (m *Manager) Snapshot() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// Start 校验环境后异步执行升级。
// 返回 nil 表示任务已成功启动，不代表升级完成——完成情况需轮询 Snapshot。
func (m *Manager) Start(ctx context.Context) error {
	c := Detect()
	if !c.CanUpdate {
		return ErrNotSupported
	}

	release, err := version.FetchLatestRelease(ctx)
	if err != nil {
		return err
	}
	hasUpdate, _ := version.IsNewerVersion(release.TagName, version.Version)
	if !hasUpdate {
		return ErrNoUpdateAvailable
	}

	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return ErrUpdateInProgress
	}
	m.running = true
	startedAt := m.now()
	m.state = State{
		Status:        StatusRunning,
		Stage:         StageDownloading,
		TargetVersion: release.TagName,
		StartedAt:     &startedAt,
	}
	m.mu.Unlock()

	// 不继承请求 ctx：HTTP 响应一返回请求就被取消，会把下载一并掐掉。
	// 用独立 ctx 并以 downloadTimeout 兜底。
	go m.run(release)
	return nil
}

func (m *Manager) run(release *version.Release) {
	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	err := m.updater.Apply(ctx, release, m.report)

	m.mu.Lock()
	defer m.mu.Unlock()
	finishedAt := m.now()
	m.state.FinishedAt = &finishedAt
	m.running = false
	if err != nil {
		m.state.Status = StatusFailed
		m.state.Error = err.Error()
		return
	}
	m.state.Status = StatusSucceeded
	m.state.Stage = StageDone
	m.state.Percent = 100
	m.state.RestartRequired = true
}

func (m *Manager) report(stage Stage, percent int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.Stage = stage
	m.state.Percent = percent
}

// Rollback 还原上一版本二进制。升级任务执行期间拒绝回滚，避免与替换过程交错。
func (m *Manager) Rollback() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return ErrUpdateInProgress
	}
	m.mu.Unlock()

	if err := m.updater.Rollback(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.RestartRequired = true
	return nil
}

// Running 是否有升级任务在执行
func (m *Manager) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}
