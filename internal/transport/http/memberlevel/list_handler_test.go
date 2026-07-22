package memberlevelhttp

import (
	"testing"

	"github.com/dujiao-next/internal/models"
)

func TestResolveLocalizedJSONPrefersLocaleThenDefault(t *testing.T) {
	m := models.JSON{
		"en-US": "Gold",
		"zh-CN": "黄金",
	}
	if got := resolveLocalizedJSON(m, "en-US", "zh-CN"); got != "Gold" {
		t.Fatalf("locale prefer: got %q", got)
	}
	if got := resolveLocalizedJSON(m, "ja-JP", "zh-CN"); got != "黄金" {
		t.Fatalf("default prefer: got %q", got)
	}
	if got := resolveLocalizedJSON(models.JSON{}, "zh-CN", "zh-CN"); got != "" {
		t.Fatalf("empty map: got %q", got)
	}
}
