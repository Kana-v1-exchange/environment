package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
)

type RedisSettings struct {
	Host     string
	Port     string
	Password string
}

type RedisHandler interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Remove(keys ...string) error

	AddOperation(currency string, price int) error
	GetOrUpdateUserToken(userID uint64, expiresAt *time.Time) (time.Time, error)
}

type redisClient struct {
	client *redis.Client
}

func (rs *RedisSettings) Connect() RedisHandler {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", rs.Host, rs.Port),
		Password: rs.Password,
		DB:       0,
	})

	status := rdb.Ping(context.Background())
	if status.Err() != nil {
		panic(fmt.Sprintf("cannot connect to the redis server %v", status.Err()))
	}

	return &redisClient{client: rdb}
}

func (rc *redisClient) Set(key string, value string) error {
	err := rc.client.Set(context.Background(), key, value, 0).Err()

	if err != nil {
		return fmt.Errorf("redis cannot set value(%v) with key (%v); err: %v", value, key, err)
	}

	return nil
}

func (rc *redisClient) Get(key string) (string, error) {
	val, err := rc.client.Get(context.Background(), key).Result()

	if err != nil {
		return "", fmt.Errorf("redis cannot return value with key %v; err: %v", key, err)
	}

	return val, nil
}

func (rc *redisClient) Remove(keys ...string) error {
	err := rc.client.Del(context.Background(), keys...).Err()
	if err != nil {
		return fmt.Errorf("redis cannot delete keys %v; err: %v", keys, err)
	}

	return nil
}

func (rc *redisClient) AddOperation(currency string, price int) error {
	err := rc.client.LPush(
		context.Background(),
		currency+RedisCurrencyOperationsSuffix,
		price,
	)

	if err != nil {
		return fmt.Errorf("cannot insert price (%v) of the currency(%v); err: %v", price, currency, err)
	}

	return nil
}

func (rc *redisClient) GetOrUpdateUserToken(userID uint64, expiresAt *time.Time) (time.Time, error) {
	curTime, err := rc.Get(fmt.Sprintf("%v%v", userID, UserTokenSuffix))
	if err != nil {
		return time.Now(), fmt.Errorf("cannot get user's (id = %v) expiresAt time; err: %v", userID, err)
	}

	if expiresAt != nil {
		err = rc.Set(fmt.Sprintf("%v%v", userID, UserTokenSuffix), expiresAt.String())
		if err != nil {
			return time.Now(), fmt.Errorf("cannot set user's (id = %v) expiresAt time; err: %v", userID, err)
		}
	}

	curTokenExpiresAt, err := time.Parse(time.RFC3339, curTime)
	if err != nil {
		return time.Now(), fmt.Errorf("cannot parse time %v as time.RFC3339; err: %v", curTime, err)
	}

	return curTokenExpiresAt, nil
}
