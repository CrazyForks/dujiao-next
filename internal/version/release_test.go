package version

import "testing"

func TestIsNewerVersion(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v1.0.1", "v1.0.0", true},
		{"v1.1.0", "v1.0.9", true},
		{"v2.0.0", "v1.99.99", true},
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.0", "v1.0.1", false},
		{"1.2.3", "v1.2.3", false},
		{"v1.2.3-rc.1", "v1.2.3", false},
		{"v1.2.4", "v1.2.3-rc.1", true},
		// 双端无法解析时回退到字符串不等比较
		{"foo", "bar", true},
		{"foo", "foo", false},
		{"", "v1.0.0", false},
	}

	for _, c := range cases {
		got, _ := IsNewerVersion(c.latest, c.current)
		if got != c.want {
			t.Errorf("IsNewerVersion(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}
