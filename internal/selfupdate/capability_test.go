package selfupdate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dujiao-next/internal/version"
)

// TestDetectBlocksSourceBuild 本地构建（BuildType=source）必须被阻断，
// 否则 go run 起来的开发进程会把自己的二进制换成线上发行版。
func TestDetectBlocksSourceBuild(t *testing.T) {
	restore := version.BuildType
	version.BuildType = version.BuildTypeSource
	t.Cleanup(func() { version.BuildType = restore })

	c := Detect()
	if c.CanUpdate {
		t.Error("source build should not be allowed to self-update")
	}
	if c.BlockReason == BlockNone {
		t.Error("blocked capability must carry a reason code")
	}
	// 容器内跑测试时会先命中容器判定，两种原因码都算正确
	if c.BlockReason != BlockSourceBuild && c.BlockReason != BlockContainer {
		t.Errorf("BlockReason = %q, want source_build (or container in CI)", c.BlockReason)
	}
}

// TestDetectReleaseBuild release 产物在非容器环境下应当放行，
// 并带上平台归档名与可执行文件路径。
func TestDetectReleaseBuild(t *testing.T) {
	if InContainer() {
		t.Skip("running inside a container; self-update is intentionally blocked there")
	}

	restore := version.BuildType
	version.BuildType = version.BuildTypeRelease
	t.Cleanup(func() { version.BuildType = restore })

	c := Detect()
	if !c.CanUpdate {
		t.Fatalf("release build should be updatable, blocked by %q", c.BlockReason)
	}
	if c.ExecPath == "" {
		t.Error("ExecPath should be populated when update is allowed")
	}
	if c.AssetName != PlatformAssetSuffix() {
		t.Errorf("AssetName = %q, want %q", c.AssetName, PlatformAssetSuffix())
	}
	if c.Deployment != DeploymentBinary {
		t.Errorf("Deployment = %q, want binary", c.Deployment)
	}
	// 没有 systemd 托管时不能自动重启——退出即停服
	if c.Supervisor == SupervisorNone && c.CanRestart {
		t.Error("CanRestart must be false without a supervisor")
	}
}

func TestDirWritable(t *testing.T) {
	dir := t.TempDir()
	if !dirWritable(dir) {
		t.Error("temp dir should be writable")
	}

	if os.Geteuid() == 0 {
		t.Skip("running as root; permission bits do not restrict access")
	}
	readonly := filepath.Join(dir, "readonly")
	if err := os.Mkdir(readonly, 0o500); err != nil {
		t.Fatal(err)
	}
	if dirWritable(readonly) {
		t.Error("0500 dir should not be writable")
	}
}

func TestBackupPath(t *testing.T) {
	if got := backupPath("/opt/dujiao/dujiao-next"); got != "/opt/dujiao/dujiao-next.backup" {
		t.Errorf("backupPath = %q", got)
	}
}
