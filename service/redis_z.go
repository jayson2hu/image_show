package service

import (
	"strconv"

	"github.com/redis/go-redis/v9"
)

func RedisZ(score int64) redis.Z {
	return redis.Z{Score: float64(score), Member: strconv.FormatInt(score, 10)}
}
