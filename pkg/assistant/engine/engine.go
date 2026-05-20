// Package engine provides a TF-IDF based rule matching engine.
// It tokenizes user input and matches it against a set of pre-defined rules,
// returning the best-matching rule and a confidence score.
package engine

import (
	"fmt"
	"math"
	"strings"
)

// MatchThreshold is the minimum TF-IDF score required for a match to be considered valid.
// Scores below this threshold result in a zero-value Rule and score 0.
const MatchThreshold = 0.1

// Tokenizer defines the interface for tokenizing text into individual tokens.
// Implementations should handle Chinese text segmentation, mixed Chinese-English
// input, and return tokens in a normalized form (lowercase, etc.).
type Tokenizer interface {
	Tokenize(input string) ([]string, error)
}

// Rule represents a single matching rule with associated intent and action.
type Rule struct {
	Name        string   // Unique identifier for the rule
	Patterns    []string // Text patterns that trigger this rule (e.g., "启动nginx容器")
	Intent      string   // Identified intent (e.g., "container_start")
	Action      string   // Associated action (e.g., "start")
	Description string   // Human-readable description of the rule
}

// doc holds pre-computed TF-IDF data for a single rule.
type doc struct {
	rule       Rule
	tf         map[string]float64
	tokenCount int
}

// Engine is the TF-IDF rule matching engine.
// It pre-computes token frequency vectors for all rules at construction time,
// then uses TF-IDF similarity to find the best matching rule for a given input.
type Engine struct {
	rules     map[string]Rule
	docs      map[string]*doc
	df        map[string]int
	n         int
	tokenizer Tokenizer
}

// NewEngine creates a new Engine with the given rules and tokenizer.
// It pre-computes TF vectors and document frequency maps for all rules.
// Rules with empty names are skipped.
func NewEngine(rules []Rule, tokenizer Tokenizer) *Engine {
	e := &Engine{
		rules:     make(map[string]Rule, len(rules)),
		docs:      make(map[string]*doc, len(rules)),
		df:        make(map[string]int),
		n:         0,
		tokenizer: tokenizer,
	}

	// First pass: tokenize all patterns and build rule docs
	for _, r := range rules {
		if r.Name == "" {
			continue
		}

		d := &doc{
			rule: r,
			tf:   make(map[string]float64),
		}

		allTokens := make(map[string]int)
		for _, pattern := range r.Patterns {
			tokens, err := tokenizer.Tokenize(pattern)
			if err != nil {
				continue
			}
			for _, token := range tokens {
				token = strings.ToLower(token)
				allTokens[token]++
			}
		}

		if len(allTokens) == 0 {
			continue
		}

		for _, count := range allTokens {
			d.tokenCount += count
		}
		for token, count := range allTokens {
			d.tf[token] = float64(count) / float64(d.tokenCount)
		}

		e.rules[r.Name] = r
		e.docs[r.Name] = d
		e.n++
	}

	// Second pass: compute document frequency (unique tokens per rule)
	for _, d := range e.docs {
		for token := range d.tf {
			e.df[token]++
		}
	}

	return e
}

// idf computes the Inverse Document Frequency for a term.
// Uses log(N/df) with Laplace smoothing (+1 to denominator) to avoid division by zero.
// Returns 0 if the term appears in all rules.
func (e *Engine) idf(term string) float64 {
	if e.n == 0 {
		return 0
	}

	count, exists := e.df[term]
	if !exists || count == 0 {
		// Term not in any rule, but could appear in query — give it a small IDF
		return math.Log(float64(e.n+1) / 1)
	}

	if count >= e.n {
		// Term appears in every rule — provides no discriminative power
		return 0
	}

	return math.Log(float64(e.n) / float64(count))
}

// Match finds the best matching rule for the given input string.
// It tokenizes the input, computes TF-IDF scores against all rules,
// and returns the highest-scoring rule above MatchThreshold.
//
// Returns:
//   - The best matching Rule (zero-value if none above threshold)
//   - A confidence score in [0, 1] (0 if no match)
//
// An error is returned only if tokenization fails.
func (e *Engine) Match(input string) (Rule, float64, error) {
	if strings.TrimSpace(input) == "" {
		return Rule{}, 0, nil
	}

	tokens, err := e.tokenizer.Tokenize(input)
	if err != nil {
		return Rule{}, 0, fmt.Errorf("Engine - Match - tokenize: %w", err)
	}

	if len(tokens) == 0 {
		return Rule{}, 0, nil
	}

	// Normalize input tokens to lowercase
	queryTokens := make([]string, len(tokens))
	for i, t := range tokens {
		queryTokens[i] = strings.ToLower(t)
	}

	// Find the best matching rule
	var (
		bestRule  Rule
		bestScore float64
	)

	for _, d := range e.docs {
		score := e.score(queryTokens, d)

		if score > bestScore {
			bestScore = score
			bestRule = d.rule
		}
	}

	if bestScore >= MatchThreshold {
		return bestRule, bestScore, nil
	}

	return Rule{}, 0, nil
}

// score computes the TF-IDF similarity score between query tokens and a rule document.
//
// Formula: score = sum(TF(token in doc) * IDF(token)) / matchingCount
// where the sum is over query tokens that also appear in the document,
// and matchingCount is the number of such tokens (minimum 1).
//
// Normalizing by matching token count (rather than total rule token count)
// ensures rules with many pattern variations are not unfairly penalized.
func (e *Engine) score(queryTokens []string, d *doc) float64 {
	if d.tokenCount == 0 || e.n == 0 {
		return 0
	}

	var sum float64
	var matched int

	for _, token := range queryTokens {
		tf, ok := d.tf[token]
		if !ok {
			continue
		}

		idf := e.idf(token)
		sum += tf * idf
		matched++
	}

	if matched == 0 {
		return 0
	}

	return sum / float64(matched)
}
