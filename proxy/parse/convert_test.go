package parse

import (
	"testing"
)

func TestIsValidURIScheme(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"vmess", true},
		{"vless", true},
		{"hysteria2", true},
		{"HTTP", true},
		{"custom+proto-1.0", true},
		{"", false},
		{"1vmess", false},
		{"foo bar", false},
		{"foo|bar", false},
		{"foo#bar", false},
		{"کانفیگ جدید vmess | پرسرعت و جدید🚩 کشور : Germanyvmess", false},
	}

	for _, tt := range tests {
		if got := isValidURIScheme(tt.in); got != tt.want {
			t.Errorf("isValidURIScheme(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestStripInlineTrailingMeta(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		wantCleaned string
		wantMeta    string
	}{
		{
			name:        "trailing pipe metadata without fragment",
			in:          "http://1.2.3.4:8080 | AL | 2631ms | 3d",
			wantCleaned: "http://1.2.3.4:8080",
			wantMeta:    "AL | 2631ms | 3d",
		},
		{
			name:        "trailing metadata with fragment",
			in:          "vless://uuid@1.2.3.4:443 | AL #US",
			wantCleaned: "vless://uuid@1.2.3.4:443#US",
			wantMeta:    "AL",
		},
		{
			name:        "spaces and pipes inside fragment are preserved",
			in:          "vmess://abc#Node | Name",
			wantCleaned: "vmess://abc#Node | Name",
			wantMeta:    "",
		},
		{
			name:        "dirty prefix before scheme does not strip scheme",
			in:          "کانفیگ جدید vmess | پرسرعت و جدید🚩 کشور : Germanyvmess://eyJhZ | tag",
			wantCleaned: "کانفیگ جدید vmess | پرسرعت و جدید🚩 کشور : Germanyvmess://eyJhZ",
			wantMeta:    "tag",
		},
		{
			name:        "dirty prefix without trailing meta stays intact",
			in:          "کانفیگ جدید vmess | پرسرعت و جدید🚩 کشور : Germanyvmess://eyJhZ",
			wantCleaned: "کانفیگ جدید vmess | پرسرعت و جدید🚩 کشور : Germanyvmess://eyJhZ",
			wantMeta:    "",
		},
		{
			name:        "empty body after scheme",
			in:          "vmess:// | tag",
			wantCleaned: "vmess://",
			wantMeta:    "tag",
		},
		{
			name:        "plain host port with trailing meta",
			in:          "1.2.3.4:8080 | AL | 100ms",
			wantCleaned: "1.2.3.4:8080",
			wantMeta:    "AL | 100ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCleaned, gotMeta := stripInlineTrailingMeta(tt.in)
			if gotCleaned != tt.wantCleaned || gotMeta != tt.wantMeta {
				t.Errorf("stripInlineTrailingMeta(%q) = (%q, %q), want (%q, %q)",
					tt.in, gotCleaned, gotMeta, tt.wantCleaned, tt.wantMeta)
			}
		})
	}
}

func TestParseProxyLinksAndConvert_DirtyLinksDoNotPanic(t *testing.T) {
	// 包含原故障复现行（schemeIdx=82, 原 cleaned 长度 27）及其他边界脏数据，
	// 混排正常链接时应跳过脏行、绝不 panic，且不影响正常节点的解析结果。
	links := []string{
		"کانفیگ جدید vmess | پرسرعت و جدید🚩 کشور : Germanyvmess://eyJhZCI6IjEuMi4zLjQifQ==",
		"prefix text | extra vless://uuid@1.2.3.4:443",
		"foo#bar://baz",
		"://missing-scheme",
		"http ://1.2.3.4:8080 | tag",
		"vmess:// | tag",
		"http://user:pass@1.2.3.4:8080 | US | 100ms",
		"snell://testpsk@2.3.4.5:8388?version=4#SnellNode",
	}

	nodes, _ := ParseProxyLinksAndConvert(links, "")
	if len(nodes) != 2 {
		t.Fatalf("ParseProxyLinksAndConvert() returned %d nodes, want 2: %+v", len(nodes), nodes)
	}

	seenTypes := map[string]bool{}
	for _, n := range nodes {
		typ, _ := n["type"].(string)
		seenTypes[typ] = true
	}
	if !seenTypes["http"] || !seenTypes["snell"] {
		t.Errorf("unexpected parsed node types: %+v", nodes)
	}
}
