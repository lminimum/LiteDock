package entity

import "time"

type Container struct {
	ID        string    `json:"id"`
	MachineID string    `json:"machine_id"`
	Name      string    `json:"name"`
	Image     string    `json:"image"`
	Status    string    `json:"status"`
	Ports     []string  `json:"ports"`
	Created   int64     `json:"created"`
	CachedAt  time.Time `json:"cached_at"`
}
