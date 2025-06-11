package postgres

import "time"

type option func(*Postgres)

func SetMaxPoolSize(size int) option {
	return func(p *Postgres) {
		p.maxPoolSize = size
	}
}

func SetMaxConnAttempts(attempts int) option {
	return func(p *Postgres) {
		p.maxConnAttempts = attempts
	}
}

func SetMaxConnTimeout(timeout time.Duration) option {
	return func(p *Postgres) {
		p.maxConnTimeout = timeout
	}
}
