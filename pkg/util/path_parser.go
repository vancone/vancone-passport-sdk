package util

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type PathPattern struct {
	Pattern      string
	RegexPattern *regexp.Regexp
	Variables    []string
	Score        int
}

type PathPatternParser struct {
	patterns []*PathPattern
}

func NewPathPatternParser() *PathPatternParser {
	return &PathPatternParser{
		patterns: make([]*PathPattern, 0),
	}
}

func (p *PathPatternParser) Parse(pattern string) (*PathPattern, error) {
	// 清理pattern
	pattern = cleanPattern(pattern)

	// 提取变量名
	vars := extractVariables(pattern)

	// 构建正则表达式
	regexPattern := pattern

	// 处理 /** 模式
	regexPattern = strings.ReplaceAll(regexPattern, "/**", "(?:/.*)?")

	// 处理 /* 模式
	regexPattern = strings.ReplaceAll(regexPattern, "/*", "/[^/]*")

	// 处理 {variable} 模式
	for _, v := range vars {
		regexPattern = strings.ReplaceAll(regexPattern,
			fmt.Sprintf("{%s}", v),
			"([^/]+)")
	}

	// 确保完全匹配
	regexPattern = "^" + regexPattern + "$"

	// 编译正则表达式
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %s: %v", pattern, err)
	}

	return &PathPattern{
		Pattern:      pattern,
		RegexPattern: regex,
		Variables:    vars,
		Score:        calculateScore(pattern),
	}, nil
}

func (p *PathPatternParser) AddPattern(pattern string) error {
	pp, err := p.Parse(pattern)
	if err != nil {
		return err
	}
	p.patterns = append(p.patterns, pp)
	// 按得分排序，具体的模式优先
	sort.Slice(p.patterns, func(i, j int) bool {
		return p.patterns[i].Score > p.patterns[j].Score
	})
	return nil
}

// Match 查找最佳匹配的模式
func (p *PathPatternParser) Match(path string) (*PathPattern, map[string]string) {
	path = cleanPattern(path)

	for _, pattern := range p.patterns {
		if matches := pattern.RegexPattern.FindStringSubmatch(path); matches != nil {
			vars := make(map[string]string)
			for i, name := range pattern.Variables {
				vars[name] = matches[i+1]
			}
			return pattern, vars
		}
	}
	return nil, nil
}

// 辅助函数
func cleanPattern(pattern string) string {
	// 移除多余的斜杠
	pattern = regexp.MustCompile(`/+`).ReplaceAllString(pattern, "/")
	// 确保以/开头
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
	// 移除末尾的/（除非是根路径）
	if len(pattern) > 1 && strings.HasSuffix(pattern, "/") {
		pattern = pattern[:len(pattern)-1]
	}
	return pattern
}

func extractVariables(pattern string) []string {
	re := regexp.MustCompile(`\{([^/]+?)\}`)
	matches := re.FindAllStringSubmatch(pattern, -1)
	vars := make([]string, len(matches))
	for i, match := range matches {
		vars[i] = match[1]
	}
	return vars
}

func calculateScore(pattern string) int {
	score := 100

	// 静态段得分高
	score += strings.Count(pattern, "/") * 10

	// 变量降低得分
	score -= strings.Count(pattern, "{") * 5

	// 通配符大幅降低得分
	score -= strings.Count(pattern, "*") * 10

	// /** 模式得分最低
	score -= strings.Count(pattern, "/**") * 20

	return score
}

/* 使用案例
func main() {
    parser := NewPathPatternParser()

    // 添加一些模式
    patterns := []string{
        "/users/{id}",
        "/users/{id}/posts/{postId}",
        "/api/v1/**",
        "/static/*",
        "/products/{category}/{id}",
    }

    for _, p := range patterns {
        if err := parser.AddPattern(p); err != nil {
            log.Printf("Error adding pattern %s: %v", p, err)
        }
    }

    // 测试匹配
    testPaths := []string{
        "/users/123",
        "/users/456/posts/789",
        "/api/v1/any/path/here",
        "/static/style.css",
        "/products/electronics/001",
    }

    for _, path := range testPaths {
        if pattern, vars := parser.Match(path); pattern != nil {
            fmt.Printf("Path: %s\n", path)
            fmt.Printf("Matched pattern: %s\n", pattern.Pattern)
            fmt.Printf("Variables: %v\n\n", vars)
        } else {
            fmt.Printf("No match found for path: %s\n\n", path)
        }
    }
}
*/
