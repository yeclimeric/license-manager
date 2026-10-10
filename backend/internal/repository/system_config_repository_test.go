package repository

import "testing"

// 白名单为空表示不限制来源
func TestIPWhitelistEmpty(t *testing.T) {
	if !IPWhitelistMatches("", "203.0.113.5") {
		t.Error("白名单为空时应放行任意 IP")
	}
	if !IPWhitelistMatches("   ", "203.0.113.5") {
		t.Error("白名单为空白字符时应放行任意 IP")
	}
}

func TestIPWhitelistExactMatch(t *testing.T) {
	if !IPWhitelistMatches("203.0.113.5", "203.0.113.5") {
		t.Error("精确匹配应当放行")
	}
	if IPWhitelistMatches("203.0.113.5", "203.0.113.6") {
		t.Error("不同 IP 不应放行")
	}
}

// 逗号分隔的多个条目
func TestIPWhitelistMultiple(t *testing.T) {
	allowed := "203.0.113.5, 198.51.100.7 ,192.0.2.1"
	for _, ip := range []string{"203.0.113.5", "198.51.100.7", "192.0.2.1"} {
		if !IPWhitelistMatches(allowed, ip) {
			t.Errorf("IP %s 应在白名单内", ip)
		}
	}
	if IPWhitelistMatches(allowed, "10.0.0.1") {
		t.Error("白名单外的 IP 不应放行")
	}
}

// CIDR 网段：白名单限定网段时网段内应放行
func TestIPWhitelistCIDR(t *testing.T) {
	allowed := "203.0.113.0/24"
	if !IPWhitelistMatches(allowed, "203.0.113.99") {
		t.Error("同网段 IP 应放行")
	}
	if IPWhitelistMatches(allowed, "203.0.114.1") {
		t.Error("不同网段不应放行")
	}
}

// 非法 CIDR 不应放行任何流量
func TestIPWhitelistInvalidCIDR(t *testing.T) {
	if IPWhitelistMatches("not-a-cidr", "203.0.113.5") {
		t.Error("非法 CIDR 不应放行")
	}
}

// 非法 IP 不得匹配
func TestIPWhitelistInvalidIP(t *testing.T) {
	if IPWhitelistMatches("203.0.113.0/24", "invalid-ip") {
		t.Error("非法 IP 不应匹配")
	}
}

// API Key 格式：sk- 前缀 + 32 位十六进制
func TestGenerateAPIKeyFormat(t *testing.T) {
	plaintext, prefix, hash, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	if len(plaintext) != 3+32 {
		t.Errorf("明文长度 = %d, want %d", len(plaintext), 3+32)
	}
	if plaintext[:3] != "sk-" {
		t.Errorf("前缀 = %q, want sk-", plaintext[:3])
	}
	if prefix != plaintext[:11] {
		t.Errorf("展示前缀 = %q, 应为明文前 11 位", prefix)
	}
	if len(hash) != 64 {
		t.Errorf("哈希长度 = %d, want 64", len(hash))
	}

	// 明文不得出现在哈希中
	if hash == plaintext {
		t.Error("哈希不应等于明文")
	}
}

// 两次生成必须不同
func TestGenerateAPIKeyUnique(t *testing.T) {
	a, _, hashA, _ := GenerateAPIKey()
	b, _, hashB, _ := GenerateAPIKey()
	if a == b {
		t.Error("两次生成的明文不应相同")
	}
	if hashA == hashB {
		t.Error("两次生成的哈希不应相同")
	}
}

// 逗号分隔字符串解析
func TestSplitAndTrim(t *testing.T) {
	if got := splitAndTrim(""); got != nil {
		t.Errorf("空串应返回 nil, got %v", got)
	}
	got := splitAndTrim(" a , b ,, c ")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("解析结果 = %v, 期望 [a b c]", got)
	}
}

// dedupeStrings 去重并保持首次出现的顺序。
//
// 角色名去重关系到是否会重复关联：同一角色提交两次不应产生两条关联记录。
func TestDedupeStrings(t *testing.T) {
	cases := []struct {
		name  string
		input []string
		want  []string
	}{
		{name: "空输入", input: nil, want: []string{}},
		{name: "无重复", input: []string{"a", "b", "c"}, want: []string{"a", "b", "c"}},
		{name: "相邻重复", input: []string{"a", "a", "b"}, want: []string{"a", "b"}},
		{name: "非相邻重复保留首次位置", input: []string{"a", "b", "a"}, want: []string{"a", "b"}},
		{name: "空白项被剔除", input: []string{"a", "", "b"}, want: []string{"a", "b"}},
		{name: "纯空白项被剔除", input: []string{"  ", "a"}, want: []string{"a"}},
		{name: "前后空白被裁剪", input: []string{" a ", "b"}, want: []string{"a", "b"}},
		{name: "裁剪后相同视为重复", input: []string{"a", " a "}, want: []string{"a"}},
		{name: "全为空白", input: []string{"", "  "}, want: []string{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dedupeStrings(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("dedupeStrings(%v) = %v, want %v", tc.input, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("第 %d 项 = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// 去重后的数量决定比对结果：len(roles) 必须等于去重后的期望数
func TestDedupeStringsDrivesRoleLookup(t *testing.T) {
	// 仓储层据此判断是否缺少角色：重复项不应被算作多个期望值
	if got := dedupeStrings([]string{"viewer", "viewer"}); len(got) != 1 {
		t.Errorf("重复角色名去重后应为 1 项, got %d", len(got))
	}
	if got := dedupeStrings([]string{"viewer", "operator"}); len(got) != 2 {
		t.Errorf("不同角色名应保留 2 项, got %d", len(got))
	}
}
