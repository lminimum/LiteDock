package postgres

import "testing"

func TestRewritePlaceholders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "no placeholders",
			query:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "single placeholder",
			query:    "SELECT * FROM users WHERE id = ?",
			expected: "SELECT * FROM users WHERE id = $1",
		},
		{
			name:     "multiple placeholders",
			query:    "INSERT INTO users(id, name, email) VALUES (?, ?, ?)",
			expected: "INSERT INTO users(id, name, email) VALUES ($1, $2, $3)",
		},
		{
			name:     "placeholder in string literal",
			query:    "SELECT * FROM users WHERE name LIKE '%?%'",
			expected: "SELECT * FROM users WHERE name LIKE '%?%'",
		},
		{
			name:     "escaped quote in string",
			query:    "SELECT * FROM users WHERE name = 'O''Brien'",
			expected: "SELECT * FROM users WHERE name = 'O''Brien'",
		},
		{
			name:     "mixed placeholders and string",
			query:    "SELECT * FROM users WHERE name = ? AND status = 'active' AND id = ?",
			expected: "SELECT * FROM users WHERE name = $1 AND status = 'active' AND id = $2",
		},
		{
			name:     "zero placeholders with string literal",
			query:    "SELECT * FROM users WHERE name = 'test'",
			expected: "SELECT * FROM users WHERE name = 'test'",
		},
		{
			name:     "many placeholders",
			query:    "INSERT INTO users(a,b,c,d,e,f,g,h,i,j) VALUES (?,?,?,?,?,?,?,?,?,?)",
			expected: "INSERT INTO users(a,b,c,d,e,f,g,h,i,j) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := rewritePlaceholders(tt.query)
			if result != tt.expected {
				t.Errorf("rewritePlaceholders(%q) = %q, want %q", tt.query, result, tt.expected)
			}
		})
	}
}
