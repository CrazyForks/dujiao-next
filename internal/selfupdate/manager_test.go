package selfupdate

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/dujiao-next/internal/version"
)

// TestManagerStartRejectsUnsupportedEnv 环境不支持时必须在发起任何网络请求前拒绝。
func TestManagerStartRejectsUnsupportedEnv(t *testing.T) {
	restore := version.BuildType
	version.BuildType = version.BuildTypeSource
	t.Cleanup(func() { version.BuildType = restore })

	m := NewManager()
	err := m.Start(context.Background())
	if !errors.Is(err, ErrNotSupported) {
		t.Fatalf("Start error = %v, want ErrNotSupported", err)
	}
	if m.Running() {
		t.Error("no task should be running after a rejected start")
	}
	if got := m.Snapshot().Status; got != StatusIdle {
		t.Errorf("status = %q, want idle", got)
	}
}

func TestManagerInitialSnapshot(t *testing.T) {
	m := NewManager()
	s := m.Snapshot()
	if s.Status != StatusIdle || s.Stage != StageIdle {
		t.Errorf("initial snapshot = %+v, want idle/idle", s)
	}
	if s.RestartRequired {
		t.Error("fresh manager should not require restart")
	}
}

// TestManagerRollbackWithoutBackup 没有备份时回滚要给出明确错误，而不是静默成功。
func TestManagerRollbackWithoutBackup(t *testing.T) {
	execPath, err := ExecutablePath()
	if err != nil {
		t.Skipf("cannot resolve executable path: %v", err)
	}
	if _, err := os.Stat(backupPath(execPath)); err == nil {
		t.Skip("a real backup exists next to the test binary")
	}

	m := NewManager()
	if err := m.Rollback(); !errors.Is(err, ErrNoBackup) {
		t.Fatalf("Rollback error = %v, want ErrNoBackup", err)
	}
}

// TestManagerReportUpdatesProgress 进度回调要如实反映在可轮询的快照里。
func TestManagerReportUpdatesProgress(t *testing.T) {
	m := NewManager()
	m.report(StageDownloading, 42)

	s := m.Snapshot()
	if s.Stage != StageDownloading || s.Percent != 42 {
		t.Errorf("snapshot = %+v, want downloading/42", s)
	}
}
