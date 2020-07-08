package module

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

func redisConnect(host, port string) (redis.Conn, error) {
	c, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}
	return c, nil
}

func RedisSet(key, value string) error {
	con, err := redisConnect("127.0.0.1", "6379")
	if err != nil {
		return err
	}
	defer con.Close()

	if _, err := con.Do("SET", key, value); err != nil {
		return err
	}
	return nil
}

func RedisGet(key string) (string, error) {
	con, err := redisConnect("127.0.0.1", "6379")
	if err != nil {
		return "_", err
	}
	defer con.Close()
	if v, err := redis.Int(con.Do("EXISTS", key)); err == nil {
		if v == 0 {
			con.Do("SET", key, "[]")
		}
	}
	value, err := redis.String(con.Do("GET", key))
	return value, err
}
