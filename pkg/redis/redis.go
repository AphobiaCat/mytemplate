package redis

import (
	"context"
	"crypto/tls"
	"mytemplate/pkg/log"
	"sync"

	"github.com/redis/go-redis/v9"

	"time"

	"mytemplate/internal/global"
	"mytemplate/pkg/util"
)

var redisManager RedisManager

type RedisManager struct {
	rdb *redis.Client
	ctx context.Context

	valuelock          []sync.Mutex
	valuelockindex     map[string]int
	valuelockindexlock sync.Mutex
}

func (rm *RedisManager) SetValue(valuekey string, value interface{}) {

	rm.valuelockindexlock.Lock()

	val, exist := rm.valuelockindex[valuekey]

	if !exist {
		rm.valuelock = append(rm.valuelock, sync.Mutex{})
		rm.valuelockindex[valuekey] = len(rm.valuelock) - 1
		val = rm.valuelockindex[valuekey]
	}

	rm.valuelockindexlock.Unlock()

	rm.valuelock[val].Lock()

	err := rm.rdb.Set(rm.ctx, valuekey, value, 0).Err()
	if err != nil {
		rm.valuelock[val].Unlock()
		log.DebugError("set value failed", err)
		return
	}

	rm.valuelock[val].Unlock()
}

func (rm *RedisManager) GetValue(valuekey string) interface{} {

	rm.valuelockindexlock.Lock()

	val, exist := rm.valuelockindex[valuekey]

	if !exist {
		rm.valuelock = append(rm.valuelock, sync.Mutex{})
		rm.valuelockindex[valuekey] = len(rm.valuelock) - 1
		val = rm.valuelockindex[valuekey]
	}

	rm.valuelockindexlock.Unlock()

	rm.valuelock[val].Lock()

	retval, err := rm.rdb.Get(rm.ctx, valuekey).Result()
	if err != nil {
		rm.valuelock[val].Unlock()
		log.DebugError("get value failed", err)
		return retval
	}
	//log.DebugLog("key value:", val)

	rm.valuelock[val].Unlock()

	return retval
}

func (rm *RedisManager) ReturnValue(valuekey string, value interface{}) {
	rm.valuelockindexlock.Lock()

	val, exist := rm.valuelockindex[valuekey]

	if !exist {
		rm.valuelockindexlock.Unlock()
		log.DebugError("ReturnValue value failed, this value no Borrow")
		return
	}

	rm.valuelockindexlock.Unlock()

	err := rm.rdb.Set(rm.ctx, valuekey, value, 0).Err()
	if err != nil {
		rm.valuelock[val].Unlock()
		log.DebugError("set value failed", err)
		return
	}

	rm.valuelock[val].Unlock()
}

func (rm *RedisManager) BorrowValue(valuekey string) interface{} {
	rm.valuelockindexlock.Lock()

	val, exist := rm.valuelockindex[valuekey]

	if !exist {
		rm.valuelock = append(rm.valuelock, sync.Mutex{})
		rm.valuelockindex[valuekey] = len(rm.valuelock) - 1
		val = rm.valuelockindex[valuekey]
	}

	rm.valuelockindexlock.Unlock()

	rm.valuelock[val].Lock()

	retval, err := rm.rdb.Get(rm.ctx, valuekey).Result()
	if err != nil {
		rm.valuelock[val].Unlock()
		log.DebugError("get value failed", err)
		return retval
	}

	return retval
}

func (rm *RedisManager) LPUSH(rediskey string, data string) {
	err := rm.rdb.LPush(rm.ctx, rediskey, data).Err()

	if err != nil {
		log.DebugError("queue set value failed", err)
	}
}

func (rm *RedisManager) QueueSet(rediskey string, data interface{}) {
	err := rm.rdb.LPush(rm.ctx, rediskey, util.BuildJson(data)).Err()

	if err != nil {
		log.DebugError("queue set value failed", err)
	}
}

func (rm *RedisManager) QueueGet(rediskey string) (string, bool) {
	task, err := rm.rdb.RPop(rm.ctx, rediskey).Result()

	if err != nil {
		if err != redis.Nil {
			log.DebugError("queue get value failed", err)
		}
		return "", false
	}
	return task, true
}

func (rm *RedisManager) StackSet(rediskey string, data interface{}) {
	err := rm.rdb.LPush(rm.ctx, rediskey, util.BuildJson(data)).Err()

	if err != nil {
		log.DebugError("stack set value failed", err)
	}
}

func (rm *RedisManager) StackGet(rediskey string) (string, bool) {
	task, err := rm.rdb.LPop(rm.ctx, rediskey).Result()

	if err != nil {
		if err != redis.Nil {
			log.DebugError("stack get value failed", err)
		}
		return "", false
	}
	return task, true
}

func (rm *RedisManager) ListRange(rediskey string, startpos int64, endpos int64) ([]string, bool) {
	values, err := rm.rdb.LRange(rm.ctx, rediskey, startpos, endpos).Result()
	if err != nil {
		log.DebugError("redis list range err:", err)
		return values, false
	}

	return values, true
}

func (rm *RedisManager) AddNum(rediskey string, num int64, timeouts ...int64) (int64, bool) {
	values, err := rm.rdb.IncrBy(rm.ctx, rediskey, num).Result()
	if err != nil {
		log.DebugError("redis incr by int err:", err)
		return values, false
	}

	if len(timeouts) > 0 {
		rm.rdb.Expire(rm.ctx, rediskey, time.Duration(timeouts[0]*1000*1000*1000))
	}

	return values, true
}

func (rm *RedisManager) AddFloatNum(rediskey string, num float64, timeouts ...int64) (float64, bool) {
	values, err := rm.rdb.IncrByFloat(rm.ctx, rediskey, num).Result()
	if err != nil {
		log.DebugError("redis incr by float err:", err)
		return values, false
	}

	if len(timeouts) > 0 {
		rm.rdb.Expire(rm.ctx, rediskey, time.Duration(timeouts[0]*1000*1000*1000))
	}

	return values, true
}

func (rm *RedisManager) HSet(rediskey string, hashkey string, value interface{}) {
	err := rm.rdb.HSet(rm.ctx, rediskey, hashkey, util.BuildJson(value)).Err()
	if err != nil {
		log.DebugError("redis HSet err:", err)
	}
}

func (rm *RedisManager) HLen(rediskey string) int64 {
	lenofmap, err := rm.rdb.HLen(rm.ctx, rediskey).Result()
	if err != nil {
		log.DebugError("redis HLen err:", err)
	}

	return lenofmap
}

func (rm *RedisManager) HGetAll(rediskey string) map[string]string {
	ret, err := rm.rdb.HGetAll(rm.ctx, rediskey).Result()
	if err != nil {
		log.DebugError("redis HGetAll err:", err)
	}
	return ret
}

func (rm *RedisManager) HExists(rediskey string, hashkey string) bool {
	ret, err := rm.rdb.HExists(rm.ctx, rediskey, hashkey).Result()
	if err != nil {
		log.DebugError("redis HExist err:", err)
	}
	return ret
}

func (rm *RedisManager) HGet(rediskey string, hashkey string) string {
	ret, err := rm.rdb.HGet(rm.ctx, rediskey, hashkey).Result()
	if err != nil {
		log.DebugError("redis HGet err:", err)
	}
	return ret
}

func (rm *RedisManager) HDel(rediskey string, hashkey string) {
	err := rm.rdb.HDel(rm.ctx, rediskey, hashkey).Err()
	if err != nil {
		log.DebugError("redis HDel err:", err)
	}
}

func (rm *RedisManager) Get(rediskey string) (string, bool) {
	values, err := rm.rdb.Get(rm.ctx, rediskey).Result()
	if err != nil {
		log.DebugError("get err:", err)
		return values, false
	}

	return values, true
}

func (rm *RedisManager) Delete(rediskey string) {
	err := rm.rdb.Del(rm.ctx, rediskey)
	if err != nil {
		log.DebugError("del value failed", err)
	}
}

func SetValue(valuekey string, value interface{}) {
	redisManager.SetValue(valuekey, value)
}

func ReturnValue(valuekey string, value interface{}) {
	redisManager.ReturnValue(valuekey, value)
}

func GetValue(valuekey string) interface{} {
	return redisManager.GetValue(valuekey)
}

func BorrowValue(valuekey string) interface{} {
	return redisManager.BorrowValue(valuekey)
}

func LPUSH(rediskey string, data string) {
	redisManager.LPUSH(rediskey, data)
}

func QueueSet(rediskey string, data interface{}) {
	redisManager.QueueSet(rediskey, data)
}

func QueueGet(rediskey string) (string, bool) {
	return redisManager.QueueGet(rediskey)
}

func StackSet(rediskey string, data interface{}) {
	redisManager.StackSet(rediskey, data)
}

func StackGet(rediskey string) (string, bool) {
	return redisManager.StackGet(rediskey)
}

func ListRange(rediskey string, startpos int64, endpos int64) ([]string, bool) {
	return redisManager.ListRange(rediskey, startpos, endpos)
}

func AddNum(key string, num int64, timeouts ...int64) (int64, bool) {
	return redisManager.AddNum(key, num, timeouts...)
}

func AddFloatNum(key string, num float64, timeouts ...int64) (float64, bool) {
	return redisManager.AddFloatNum(key, num, timeouts...)
}

func HSet(rediskey string, hashkey string, value interface{}) {
	redisManager.HSet(rediskey, hashkey, value)
}

func HLen(rediskey string) int64 {
	return redisManager.HLen(rediskey)
}

func HGetAll(rediskey string) map[string]string {
	return redisManager.HGetAll(rediskey)
}

func HExists(rediskey string, hashkey string) bool {
	return redisManager.HExists(rediskey, hashkey)
}

func HGet(rediskey string, hashkey string) string {
	return redisManager.HGet(rediskey, hashkey)
}

func HDel(rediskey string, hashkey string) {
	redisManager.HDel(rediskey, hashkey)
}

func Get(key string) (string, bool) {
	return redisManager.Get(key)
}

func Delete(key string) {
	redisManager.Delete(key)
}

func init() {

	redisManager.valuelockindex = make(map[string]int)

	redisManager.ctx = context.Background()

	if global.AppConfig.Redis.EnableTls {
		redisManager.rdb = redis.NewClient(&redis.Options{
			Addr:      global.AppConfig.Redis.Host,
			Password:  global.AppConfig.Redis.Password,
			DB:        global.AppConfig.Redis.DB,
			TLSConfig: &tls.Config{},
		})
	} else {
		redisManager.rdb = redis.NewClient(&redis.Options{
			Addr:     global.AppConfig.Redis.Host,
			Password: global.AppConfig.Redis.Password,
			DB:       global.AppConfig.Redis.DB,
		})
	}

	_, err := redisManager.rdb.Ping(redisManager.ctx).Result()
	if err != nil {
		log.DebugError("unable connet Redis:", err)
		panic(err)
	}
	log.DebugLog("connect redis server succ ip[", global.AppConfig.Redis.Host, "] db[", global.AppConfig.Redis.DB, "]")

	//rdb := redisManager.rdb
	//log.DebugLogVAR(rdb)
}

func CloseRedis() {
	redisManager.rdb.Close()
}
