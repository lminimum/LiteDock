package mysql

import "time"

type Option func(*MySQL)

func MaxOpenConns(n int) Option {
	return func(m *MySQL) {
		m.maxOpenConns = n
	}
}

func ConnMaxLifetime(d time.Duration) Option {
	return func(m *MySQL) {
		m.connMaxLifetime = d
	}
}

func ConnAttempts(n int) Option {
	return func(m *MySQL) {
		m.connAttempts = n
	}
}

func ConnTimeout(d time.Duration) Option {
	return func(m *MySQL) {
		m.connTimeout = d
	}
}
