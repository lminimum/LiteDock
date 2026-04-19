package entity

import (
	"time"
)

// AuthMethod represents the SSH authentication method for a remote machine.
type AuthMethod string

const (
	// AuthMethodPassword uses password-based SSH authentication.
	AuthMethodPassword AuthMethod = "password"
	// AuthMethodKey uses SSH key-based authentication.
	AuthMethodKey AuthMethod = "key"
)

// RemoteMachine represents a remote Docker host managed by the system.
type RemoteMachine struct {
	ID         string     `json:"id" db:"id"`
	Name       string     `json:"name" db:"name"`
	Host       string     `json:"host" db:"host"`
	Port       int        `json:"port" db:"port"`
	Username   string     `json:"username" db:"username"`
	AuthMethod AuthMethod `json:"auth_method" db:"auth_method"`
	Password   string     `json:"-" db:"password"`
	SSHKeyPath string     `json:"ssh_key_path,omitempty" db:"ssh_key_path"`
	SSHKey     string     `json:"-" db:"ssh_key"`
	DockerHost string     `json:"docker_host" db:"docker_host"`
	Status     string     `json:"status" db:"status"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}
