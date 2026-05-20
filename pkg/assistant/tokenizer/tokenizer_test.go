package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTokenizer(t *testing.T) {
	tk, err := NewTokenizer()
	require.NoError(t, err)
	require.NotNil(t, tk)
	tk.Close()
}

func TestTokenize_ChineseText(t *testing.T) {
	tk, err := NewTokenizer()
	require.NoError(t, err)
	defer tk.Close()

	tokens, err := tk.Tokenize("启动nginx容器")
	require.NoError(t, err)
	require.NotEmpty(t, tokens)

	hasNginx := false
	hasContainer := false
	for _, token := range tokens {
		if token == "nginx" {
			hasNginx = true
		}
		if token == "容器" {
			hasContainer = true
		}
	}
	require.True(t, hasNginx, "expected 'nginx' in tokens: %v", tokens)
	require.True(t, hasContainer, "expected '容器' in tokens: %v", tokens)
}

func TestTokenize_EmptyString(t *testing.T) {
	tk, err := NewTokenizer()
	require.NoError(t, err)
	defer tk.Close()

	tokens, err := tk.Tokenize("")
	require.NoError(t, err)
	require.Len(t, tokens, 0)
}

func TestTokenize_SpecialCharacters(t *testing.T) {
	tk, err := NewTokenizer()
	require.NoError(t, err)
	defer tk.Close()

	_, err = tk.Tokenize("!@#$%^&*()_+-=[]{}|;':\",./<>?")
	require.NoError(t, err)
}

func TestTokenize_MixedChineseEnglish(t *testing.T) {
	tk, err := NewTokenizer()
	require.NoError(t, err)
	defer tk.Close()

	tokens, err := tk.Tokenize("启动Docker容器并查看日志")
	require.NoError(t, err)
	require.NotEmpty(t, tokens)
}

func TestTokenizer_Close(t *testing.T) {
	tk, err := NewTokenizer()
	require.NoError(t, err)

	err = tk.Close()
	require.NoError(t, err)

	err = tk.Close()
	require.NoError(t, err)
}