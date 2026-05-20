package tokenizer

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

var defaultTerms = func() []string {
	chinese := []string{
		"启动", "开启", "运行", "开始",
		"停止", "关闭", "停掉",
		"重启", "重新启动",
		"创建", "新建", "新增",
		"删除", "移除", "销毁",
		"拉取", "下载",
		"检查", "查看", "查看日志",
		"列表", "列出", "显示", "查看列表",
		"日志", "记录", "输出",
		"状态", "统计", "资源", "监控", "性能", "资源使用",
		"详情", "信息", "详情信息",
		"网络", "网络详情",
		"卷", "存储卷", "存储",
		"使用", "设置", "配置", "管理", "修改", "更新",
		"切换", "恢复", "暂停", "终止",
		"容器",
		"镜像", "下载镜像",
		"应用",
		"连接", "断开",
		"允许", "拒绝",
		"进入", "退出",
		"保存", "加载",
		"上传", "下载",
	}

	english := []string{
		"docker", "container", "image", "network", "volume",
		"nginx", "redis", "mysql", "postgres", "postgresql", "mongo", "mongodb",
		"rabbitmq", "nats", "kafka", "elasticsearch",
		"logs", "stats", "inspect", "start", "stop", "restart",
		"delete", "remove", "create", "pull", "run", "exec",
		"compose", "swarm", "stack", "service",
		"health", "healthz", "ready", "live",
		"cpu", "memory", "disk", "io",
		"json", "yaml", "toml", "xml",
		"config", "conf", "env",
		"localhost", "http", "https", "tcp", "udp", "tls",
	}

	common := []string{
		"一个", "这个", "那个", "哪个",
		"什么", "怎么", "为什么",
		"可以", "能够", "需要",
		"已经", "正在", "将要",
		"所有", "全部", "每个",
		"根据", "按照", "通过",
		"之后", "之前", "期间",
		"关于", "对于",
		"没有", "不是", "不能",
	}

	all := make([]string, 0, len(chinese)+len(english)+len(common))
	all = append(all, chinese...)
	all = append(all, english...)
	all = append(all, common...)
	return all
}()

type Dictionary struct {
	terms  map[string]struct{}
	maxLen int
}

func NewDictionary(terms []string) *Dictionary {
	dict := &Dictionary{
		terms: make(map[string]struct{}, len(terms)),
	}
	for _, term := range terms {
		lower := strings.ToLower(term)
		dict.terms[lower] = struct{}{}
		dict.terms[term] = struct{}{}
	}
	for _, term := range terms {
		runeLen := utf8.RuneCountInString(term)
		if runeLen > dict.maxLen {
			dict.maxLen = runeLen
		}
	}
	return dict
}

type Tokenizer struct {
	dict *Dictionary
}

func NewTokenizer() (*Tokenizer, error) {
	dict := NewDictionary(defaultTerms)
	return &Tokenizer{dict: dict}, nil
}

func (t *Tokenizer) Tokenize(input string) ([]string, error) {
	if input == "" {
		return []string{}, nil
	}

	runes := []rune(input)
	length := len(runes)
	tokens := make([]string, 0, length/2)
	i := 0

	for i < length {
		maxLookup := t.dict.maxLen
		if length-i < maxLookup {
			maxLookup = length - i
		}

		matched := false
		for j := maxLookup; j >= 1; j-- {
			word := string(runes[i : i+j])
			wordLower := strings.ToLower(word)
			if _, ok := t.dict.terms[word]; ok || wordLower != word {
				if _, ok := t.dict.terms[wordLower]; ok {
					tokens = append(tokens, wordLower)
					i += j
					matched = true
					break
				}
			}
			if _, ok := t.dict.terms[word]; ok {
				tokens = append(tokens, word)
				i += j
				matched = true
				break
			}
		}
		if matched {
			continue
		}

		ch := runes[i]

		if isLatinLetter(ch) {
			start := i
			for i < length && isLatinLetter(runes[i]) {
				i++
			}
			word := strings.ToLower(string(runes[start:i]))
			tokens = append(tokens, word)
			continue
		}

		if unicode.IsDigit(ch) {
			start := i
			for i < length && unicode.IsDigit(runes[i]) {
				i++
			}
			tokens = append(tokens, string(runes[start:i]))
			continue
		}

		if unicode.Is(unicode.Han, ch) || unicode.Is(unicode.Hiragana, ch) || unicode.Is(unicode.Katakana, ch) ||
			(ch >= 0xAC00 && ch <= 0xD7AF) {
			tokens = append(tokens, string(ch))
			i++
			continue
		}

		i++
	}

	return tokens, nil
}

func (t *Tokenizer) Close() error {
	return nil
}

func isLatinLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func init() {
	sort.Strings(defaultTerms)
}
