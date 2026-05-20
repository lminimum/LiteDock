package engine

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// mockTokenizer splits input on whitespace and lowercases tokens.
type mockTokenizer struct{}

func (m *mockTokenizer) Tokenize(input string) ([]string, error) {
	if strings.TrimSpace(input) == "" {
		return nil, nil
	}
	return strings.Fields(strings.ToLower(input)), nil
}

func testRules() []Rule {
	return []Rule{
		{Name: "start", Patterns: []string{"alpha bravo"}, Intent: "alpha_bravo", Action: "do_ab", Description: "Match alpha and bravo"},
		{Name: "charlie", Patterns: []string{"alpha charlie"}, Intent: "alpha_charlie", Action: "do_ac", Description: "Match alpha and charlie"},
		{Name: "delta", Patterns: []string{"delta echo"}, Intent: "delta_echo", Action: "do_de", Description: "Match delta and echo"},
		{Name: "foxtrot", Patterns: []string{"foxtrot golf"}, Intent: "foxtrot_golf", Action: "do_fg", Description: "Match foxtrot and golf"},
	}
}

func TestNewEngine_Precomputes(t *testing.T) {
	rules := testRules()
	eng := NewEngine(rules, &mockTokenizer{})

	require.NotNil(t, eng)
	require.Equal(t, 4, eng.n)
	require.Len(t, eng.rules, 4)
	require.Len(t, eng.docs, 4)

	startDoc := eng.docs["start"]
	require.NotNil(t, startDoc)
	require.Equal(t, 2, startDoc.tokenCount)
	require.InDelta(t, 0.5, startDoc.tf["alpha"], 0.001)
	require.InDelta(t, 0.5, startDoc.tf["bravo"], 0.001)

	require.Equal(t, 2, eng.df["alpha"])
	require.Equal(t, 1, eng.df["bravo"])
}

func TestNewEngine_SkipsEmptyNamedRules(t *testing.T) {
	rules := []Rule{
		{Name: "", Patterns: []string{"alpha bravo"}, Intent: "drop"},
		{Name: "valid", Patterns: []string{"charlie"}, Intent: "keep"},
	}
	eng := NewEngine(rules, &mockTokenizer{})
	require.Equal(t, 1, eng.n)
	require.Len(t, eng.rules, 1)
	_, ok := eng.rules["valid"]
	require.True(t, ok)
}

func TestNewEngine_SkipsRulesWithNoTokens(t *testing.T) {
	// A rule whose patterns tokenize to empty (all separators, no words)
	rules := []Rule{
		{Name: "empty", Patterns: []string{"   "}, Intent: "drop"},
		{Name: "valid", Patterns: []string{"hello world"}, Intent: "keep"},
	}
	eng := NewEngine(rules, &mockTokenizer{})
	require.Equal(t, 1, eng.n)
}

func TestMatch_ExactMatch(t *testing.T) {
	rules := testRules()
	eng := NewEngine(rules, &mockTokenizer{})

	matched, score, err := eng.Match("alpha bravo")
	require.NoError(t, err)
	require.Equal(t, "start", matched.Name)
	require.Greater(t, score, 0.5)
}

func TestMatch_ExactMatchCaseInsensitive(t *testing.T) {
	rules := testRules()
	eng := NewEngine(rules, &mockTokenizer{})

	matched, score, err := eng.Match("ALPHA BRAVO")
	require.NoError(t, err)
	require.Equal(t, "start", matched.Name)
	require.Greater(t, score, 0.5)
}

func TestMatch_ExactMatchAllRules(t *testing.T) {
	rules := testRules()
	eng := NewEngine(rules, &mockTokenizer{})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"alpha_bravo", "alpha bravo", "start"},
		{"alpha_charlie", "alpha charlie", "charlie"},
		{"delta_echo", "delta echo", "delta"},
		{"foxtrot_golf", "foxtrot golf", "foxtrot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, score, err := eng.Match(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, matched.Name)
			require.Greater(t, score, 0.5)
		})
	}
}

func TestMatch_PartialMatch(t *testing.T) {
	rules := testRules()
	eng := NewEngine(rules, &mockTokenizer{})

	// "alpha" appears in rules "start" and "charlie" — lower IDF than "bravo"
	// should produce an intermediate score between no-match and exact-match
	matched, score, err := eng.Match("alpha")
	require.NoError(t, err)
	require.NotEmpty(t, matched.Name)
	require.Greater(t, score, 0.1)
	require.Less(t, score, 0.5)
}

func TestMatch_MultipleRules_PicksBest(t *testing.T) {
	rules := testRules()
	eng := NewEngine(rules, &mockTokenizer{})

	// "alpha" appears in both "start" and "charlie" — both equally scored
	// The first matching rule encountered wins (implementation detail)
	matched, score, err := eng.Match("alpha")
	require.NoError(t, err)
	require.NotEmpty(t, matched.Name)
	require.Greater(t, score, MatchThreshold)
}

func TestMatch_NoMatch(t *testing.T) {
	rules := testRules()
	eng := NewEngine(rules, &mockTokenizer{})

	matched, score, err := eng.Match("xylophone zephyr")
	require.NoError(t, err)
	require.Equal(t, "", matched.Name)
	require.Equal(t, "", matched.Intent)
	require.Equal(t, 0.0, score)
}

func TestMatch_EmptyInput(t *testing.T) {
	rules := testRules()
	eng := NewEngine(rules, &mockTokenizer{})

	matched, score, err := eng.Match("")
	require.NoError(t, err)
	require.Equal(t, "", matched.Name)
	require.Equal(t, 0.0, score)
}

func TestMatch_WhitespaceOnly(t *testing.T) {
	rules := testRules()
	eng := NewEngine(rules, &mockTokenizer{})

	matched, score, err := eng.Match("   ")
	require.NoError(t, err)
	require.Equal(t, "", matched.Name)
	require.Equal(t, 0.0, score)
}

func TestMatch_NoRules(t *testing.T) {
	eng := NewEngine(nil, &mockTokenizer{})

	matched, score, err := eng.Match("anything")
	require.NoError(t, err)
	require.Equal(t, "", matched.Name)
	require.Equal(t, 0.0, score)
}

func TestMatch_EmptyRuleList(t *testing.T) {
	eng := NewEngine([]Rule{}, &mockTokenizer{})

	matched, score, err := eng.Match("test")
	require.NoError(t, err)
	require.Equal(t, "", matched.Name)
	require.Equal(t, 0.0, score)
}

func TestMatch_SingleTokenVsSingleTokenRule(t *testing.T) {
	rules := []Rule{
		{Name: "hello", Patterns: []string{"hello"}, Intent: "greet", Action: "greet", Description: "Greeting"},
		{Name: "bye", Patterns: []string{"goodbye"}, Intent: "farewell", Action: "farewell", Description: "Farewell"},
	}
	eng := NewEngine(rules, &mockTokenizer{})

	matched, score, err := eng.Match("hello")
	require.NoError(t, err)
	require.Equal(t, "hello", matched.Name)
	require.Greater(t, score, 0.3)
}

func TestMatch_ScoreBelowThreshold(t *testing.T) {
	rules := []Rule{
		{Name: "alpha", Patterns: []string{"alpha beta gamma delta epsilon"}, Intent: "multi", Action: "multi", Description: "Multi-token"},
	}
	eng := NewEngine(rules, &mockTokenizer{})

	// "zeta" is in none of the rules — should score 0
	matched, score, err := eng.Match("zeta")
	require.NoError(t, err)
	require.Equal(t, "", matched.Name)
	require.Equal(t, 0.0, score)
}

func TestMatch_ReturnsIntentAndAction(t *testing.T) {
	rules := testRules()
	eng := NewEngine(rules, &mockTokenizer{})

	matched, score, err := eng.Match("alpha bravo")
	require.NoError(t, err)
	require.Greater(t, score, 0.5)
	require.Equal(t, "alpha_bravo", matched.Intent)
	require.Equal(t, "do_ab", matched.Action)
	require.Equal(t, "Match alpha and bravo", matched.Description)
}

func TestIDF_DiscriminativePower(t *testing.T) {
	// "rare" appears in 1 rule, "common" appears in 3 rules
	// Since "common" is in ALL rules, IDF(common)=log(3/3)=0 — no discriminative power
	rules := []Rule{
		{Name: "r1", Patterns: []string{"common rare"}, Intent: "c1", Action: "a1", Description: ""},
		{Name: "r2", Patterns: []string{"common freq"}, Intent: "c2", Action: "a2", Description: ""},
		{Name: "r3", Patterns: []string{"common freq"}, Intent: "c3", Action: "a3", Description: ""},
	}
	eng := NewEngine(rules, &mockTokenizer{})

	// "rare" appears only in r1 — should match with high confidence
	matchedRare, scoreRare, err := eng.Match("rare")
	require.NoError(t, err)
	require.Equal(t, "r1", matchedRare.Name)
	require.Greater(t, scoreRare, 0.2)

	// "common" is in every rule → IDF=0 → scores 0 for all rules → no match
	matchedCommon, scoreCommon, err := eng.Match("common")
	require.NoError(t, err)
	require.Equal(t, "", matchedCommon.Name)
	require.Equal(t, 0.0, scoreCommon)
}

func TestMatch_MultiplePatternsPerRule(t *testing.T) {
	rules := []Rule{
		{
			Name:        "multi",
			Patterns:    []string{"start container", "launch service", "begin task"},
			Intent:      "multi_start",
			Action:      "start",
			Description: "Multiple start patterns",
		},
		{
			Name:        "other",
			Patterns:    []string{"stop daemon"},
			Intent:      "other_stop",
			Action:      "stop",
			Description: "Dummy rule so IDF has discriminative power",
		},
	}
	eng := NewEngine(rules, &mockTokenizer{})

	matched, score, err := eng.Match("start container")
	require.NoError(t, err)
	require.Equal(t, "multi", matched.Name)
	require.Greater(t, score, 0.1)

	matched2, score2, err := eng.Match("launch service")
	require.NoError(t, err)
	require.Equal(t, "multi", matched2.Name)
	require.Greater(t, score2, 0.1)

	matched3, score3, err := eng.Match("start service")
	require.NoError(t, err)
	require.Equal(t, "multi", matched3.Name)
	require.Greater(t, score3, 0.0)
}
