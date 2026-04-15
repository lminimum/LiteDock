package sqlite

import "time"

type Option func(*SQLite)

func ConnMaxLifetime(d time.Duration) Option {
	return func(s *SQLite) {
		s.connMaxLifetime = d
	}
}

func ConnAttempts(n int) Option {
	return func(s *SQLite) {
		s.connAttempts = n
	}
}

func ConnTimeout(d time.Duration) Option {
	return func(s *SQLite) {
		s.connTimeout = d
	}
}
