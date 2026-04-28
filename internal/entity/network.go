package entity

import "time"

type Network struct {
	ID         string            `json:"id"`
	MachineID  string            `json:"machine_id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	Internal   bool               `json:"internal"`
	Attachable bool              `json:"attachable"`
	EnableIPv6 bool              `json:"enable_ipv6"`
	Created    string            `json:"created"`
	Labels     map[string]string `json:"labels"`
	CachedAt   time.Time         `json:"cached_at"`
}