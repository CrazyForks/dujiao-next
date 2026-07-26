package selfupdate

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dujiao-next/internal/version"
)

// TestSelectAssets 用 v1.3.1 发行版的真实附件清单验证选择逻辑：
// 里面同时存在 dujiao-all_* 归档，且校验和文件名带项目名与版本号前缀。
func TestSelectAssets(t *testing.T) {
	suffix := PlatformAssetSuffix()
	assets := []version.Asset{
		{Name: "dujiao-all_v1.3.1_Linux_arm64.tar.gz", DownloadURL: "https://github.com/all-arm"},
		{Name: "dujiao-all_v1.3.1_Linux_x86_64.tar.gz", DownloadURL: "https://github.com/all-x86"},
		{Name: "dujiao-next_1.3.1_checksums.txt", DownloadURL: "https://github.com/sums"},
		{Name: "dujiao-next_v1.3.1_Darwin_arm64.tar.gz", DownloadURL: "https://github.com/darwin-arm"},
		{Name: "dujiao-next_v1.3.1_Darwin_x86_64.tar.gz", DownloadURL: "https://github.com/darwin-x86"},
		{Name: "dujiao-next_v1.3.1_Linux_arm64.tar.gz", DownloadURL: "https://github.com/linux-arm"},
		{Name: "dujiao-next_v1.3.1_Linux_x86_64.tar.gz", DownloadURL: "https://github.com/linux-x86"},
		{Name: "dujiao-next_v1.3.1_Windows_arm64.zip", DownloadURL: "https://github.com/win-arm"},
		{Name: "dujiao-next_v1.3.1_Windows_x86_64.zip", DownloadURL: "https://github.com/win-x86"},
	}

	archiveURL, checksumURL, assetName, err := selectAssets(assets)
	if err != nil {
		t.Fatalf("selectAssets returned error for suffix %s: %v", suffix, err)
	}
	if !strings.HasSuffix(assetName, suffix) {
		t.Errorf("picked asset %q does not match platform suffix %q", assetName, suffix)
	}
	// dujiao-all_* 的平台后缀与 dujiao-next_* 完全相同，绝不能选中
	if !strings.HasPrefix(assetName, archivePrefix) {
		t.Errorf("picked asset %q is not a %s archive", assetName, archivePrefix)
	}
	if archiveURL == "" {
		t.Error("archive url is empty")
	}
	if checksumURL != "https://github.com/sums" {
		t.Errorf("checksum url = %q, want the checksums url", checksumURL)
	}
}

// TestSelectAssetsIgnoresOtherProjects 只有 dujiao-all 归档时必须报缺失，
// 而不是错误地拿别的项目的产物去替换自己。
func TestSelectAssetsIgnoresOtherProjects(t *testing.T) {
	assets := []version.Asset{
		{Name: "dujiao-all_v1.3.1_Linux_arm64.tar.gz", DownloadURL: "https://github.com/all-arm"},
		{Name: "dujiao-all_v1.3.1_Linux_x86_64.tar.gz", DownloadURL: "https://github.com/all-x86"},
		{Name: "dujiao-all_v1.3.1_Darwin_arm64.tar.gz", DownloadURL: "https://github.com/all-darwin-arm"},
		{Name: "dujiao-all_v1.3.1_Darwin_x86_64.tar.gz", DownloadURL: "https://github.com/all-darwin-x86"},
	}
	if _, _, _, err := selectAssets(assets); err == nil {
		t.Fatal("expected ErrAssetNotFound when only other projects' archives are present")
	}
}

func TestIsChecksumAsset(t *testing.T) {
	cases := map[string]bool{
		"dujiao-next_1.3.1_checksums.txt":        true, // goreleaser 默认模板
		"checksums.txt":                          true, // 显式配置成裸文件名
		"dujiao-all_1.3.1_checksums.txt":         false,
		"dujiao-next_v1.3.1_Linux_x86_64.tar.gz": false,
		"notes.txt":                              false,
	}
	for name, want := range cases {
		if got := isChecksumAsset(name); got != want {
			t.Errorf("isChecksumAsset(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestSelectAssetsMissingPlatform(t *testing.T) {
	_, _, _, err := selectAssets([]version.Asset{{Name: "checksums.txt", DownloadURL: "https://github.com/sums"}})
	if err == nil {
		t.Fatal("expected error when no platform archive is present")
	}
}

func TestValidateDownloadURL(t *testing.T) {
	cases := []struct {
		raw     string
		wantErr bool
	}{
		{"https://github.com/dujiao-next/dujiao-next/releases/download/v1/x.tar.gz", false},
		{"https://objects.githubusercontent.com/foo", false},
		{"https://release-assets.githubusercontent.com/foo", false},
		// 非 HTTPS 一律拒绝
		{"http://github.com/foo", true},
		// 第三方主机：即便 GitHub API 响应被篡改也不会下载
		{"https://evil.com/foo", true},
		// 后缀伪装成白名单域名
		{"https://github.com.evil.com/foo", true},
		{"file:///etc/passwd", true},
	}

	for _, c := range cases {
		u, err := url.Parse(c.raw)
		if err != nil {
			t.Fatalf("parse %q: %v", c.raw, err)
		}
		gotErr := validateDownloadURL(u) != nil
		if gotErr != c.wantErr {
			t.Errorf("validateDownloadURL(%q) error = %v, want error = %v", c.raw, gotErr, c.wantErr)
		}
	}
}

func TestLookupChecksum(t *testing.T) {
	content := `abc123  dujiao-next_v1.5.0_Linux_x86_64.tar.gz
def456 *dujiao-next_v1.5.0_Darwin_arm64.tar.gz
garbage-line
`
	if got := lookupChecksum(content, "dujiao-next_v1.5.0_Linux_x86_64.tar.gz"); got != "abc123" {
		t.Errorf("lookupChecksum = %q, want abc123", got)
	}
	// sha256sum 的二进制模式会给文件名加 * 前缀
	if got := lookupChecksum(content, "dujiao-next_v1.5.0_Darwin_arm64.tar.gz"); got != "def456" {
		t.Errorf("lookupChecksum with binary-mode prefix = %q, want def456", got)
	}
	if got := lookupChecksum(content, "missing.tar.gz"); got != "" {
		t.Errorf("lookupChecksum for unknown asset = %q, want empty", got)
	}
}

func TestExtractBinary(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "release.tar.gz")
	writeTarGz(t, archive, map[string]string{
		"README.md":          "readme",
		"config.yml.example": "config",
		binaryName:           "fake-binary-content",
	})

	dest := filepath.Join(dir, "extracted")
	if err := extractBinary(archive, dest); err != nil {
		t.Fatalf("extractBinary: %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if string(got) != "fake-binary-content" {
		t.Errorf("extracted content = %q, want the binary entry", got)
	}
}

func TestExtractBinaryMissingEntry(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "release.tar.gz")
	writeTarGz(t, archive, map[string]string{"README.md": "readme"})

	if err := extractBinary(archive, filepath.Join(dir, "extracted")); err == nil {
		t.Fatal("expected error when archive has no binary entry")
	}
}

// TestExtractBinaryIgnoresPathTraversal 归档内的 ../ 路径不能影响写入位置：
// 目标路径由调用方指定，只按 basename 匹配条目。
func TestExtractBinaryIgnoresPathTraversal(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "release.tar.gz")
	writeTarGz(t, archive, map[string]string{
		"../../../tmp/" + binaryName: "evil",
	})

	dest := filepath.Join(dir, "extracted")
	if err := extractBinary(archive, dest); err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("binary was not written to the requested destination: %v", err)
	}
}

func TestSwapBinaryAndRollback(t *testing.T) {
	dir := t.TempDir()
	exec := filepath.Join(dir, binaryName)
	if err := os.WriteFile(exec, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	newBin := filepath.Join(dir, "new")
	if err := os.WriteFile(newBin, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := swapBinary(exec, newBin); err != nil {
		t.Fatalf("swapBinary: %v", err)
	}

	current, _ := os.ReadFile(exec)
	if string(current) != "new" {
		t.Errorf("after swap, binary = %q, want new", current)
	}
	backup, err := os.ReadFile(backupPath(exec))
	if err != nil {
		t.Fatalf("backup missing after swap: %v", err)
	}
	if string(backup) != "old" {
		t.Errorf("backup = %q, want old", backup)
	}
}

func TestFileSHA256(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f")
	content := []byte("hello dujiao")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(content)
	want := hex.EncodeToString(sum[:])

	got, err := fileSHA256(path)
	if err != nil {
		t.Fatalf("fileSHA256: %v", err)
	}
	if got != want {
		t.Errorf("fileSHA256 = %q, want %q", got, want)
	}
}

func TestPlatformAssetSuffix(t *testing.T) {
	suffix := PlatformAssetSuffix()
	// 与 .goreleaser.yaml 的 name_template 对齐：首字母大写的 OS + 归一化后的架构
	if !strings.HasPrefix(suffix, "_") {
		t.Errorf("suffix %q should start with underscore", suffix)
	}
	if !strings.HasSuffix(suffix, ".tar.gz") && !strings.HasSuffix(suffix, ".zip") {
		t.Errorf("suffix %q has unexpected extension", suffix)
	}
	if strings.Contains(suffix, "amd64") {
		t.Errorf("suffix %q should map amd64 to x86_64", suffix)
	}
}

func writeTarGz(t *testing.T, path string, entries map[string]string) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	for name, content := range entries {
		if err := tw.WriteHeader(&tar.Header{
			Name:     name,
			Mode:     0o755,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
}
