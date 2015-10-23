package main

import (
	"github.com/garyburd/redigo/redis"
)

var (
	pool *redis.Pool
)

func NewPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     0,
		IdleTimeout: 0,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
}
