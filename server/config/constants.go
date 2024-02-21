package config

import "time"

const (
	DEFAULT_HOSTNAME                string        = "docker.localhost"
	JWT_EXPIRY_DURATION_HOURS       time.Duration = time.Hour * 48
	JWT_COOKIE_EXPIRY_DURATION_DAYS float64       = 30
)
