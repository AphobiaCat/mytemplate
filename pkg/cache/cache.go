package cache

import (
	"context"
	"crypto/tls"
	"mytemplate/pkg/log"
	"time"

	"github.com/redis/go-redis/v9"

	"mytemplate/internal/global"
	public "mytemplate/pkg/util"
)

var cacheManager CacheManager

type CacheManager struct {
	rdb *redis.Client
	ctx context.Context
}

type StandardcmCache struct {
	D    string `json:"d"`    //user data
	L    int64  `json:"l"`    //last update time
	LW   int64  `json:"lw"`   //last work time
	W    bool   `json:"w"`    //is working update status
	Wait bool   `json:"wait"` //wait frist
}

type NewCacheFunc func() (interface{}, error)

func (cm *CacheManager) Init(serverip string, password string, DB int, enabletls bool) {

	cm.ctx = context.Background()

	if enabletls {
		cm.rdb = redis.NewClient(&redis.Options{
			Addr:      serverip,
			Password:  password,
			DB:        DB,
			TLSConfig: &tls.Config{},
		})
	} else {
		cm.rdb = redis.NewClient(&redis.Options{
			Addr:     serverip,
			Password: password,
			DB:       DB,
		})
	}

	_, err := cm.rdb.Ping(cm.ctx).Result()

	if err != nil {
		log.DebugError("unable connet Redis:", err)
	} else {
		log.DebugLog("connect redis server succ")
	}
}

func (cm *CacheManager) SetCache(key string, value interface{}, configtime ...int64) {
	maxalivetime := int64(1000 * 1000 * 1000 * 60 * 2)

	if len(configtime) == 1 {
		maxalivetime = configtime[0] * 1000 * 1000 * 1000
	}

	nowtime := public.NowTimeS()

	var newCacheData StandardcmCache

	newCacheData.D = public.BuildJson(value)
	newCacheData.L = nowtime + maxalivetime //keep no update by get cache default force update time
	newCacheData.LW = 0
	newCacheData.W = false
	newCacheData.Wait = false

	err := cm.rdb.Set(cm.ctx, key, public.BuildJson(newCacheData), time.Duration(maxalivetime)).Err()
	if err != nil {
		log.DebugError("set value failed", err)
	}
}

func (cm *CacheManager) GetCache(key string, newCacheFunc NewCacheFunc, configtime ...int64) (ret string, err error) {
	ret = ""

	forceupdatetime := int64(60 * 2)
	maxworktime := int64(60 * 5)
	maxalivetime := int64(1000 * 1000 * 1000 * 60 * 10) //ns -> us -> ms -> s

	switch len(configtime) {
	case 0:
		//default config
	case 1:
		forceupdatetime = configtime[0]
	case 2:
		forceupdatetime = configtime[0]
		maxworktime = configtime[1]
	case 3:
		forceupdatetime = configtime[0]
		maxworktime = configtime[1]
		maxalivetime = configtime[2] * 1000 * 1000 * 1000
	}

	nowtime := public.NowTimeS()

	retval, err := cm.rdb.Get(cm.ctx, key).Result()

	if err != nil {
		if err == redis.Nil {
			retval = ""
		} else {
			log.DebugError("get value failed", err)
			return
		}
	}

	//log.DebugLog("retval : ", retval)

	var cacheData StandardcmCache
	public.ParserJson(retval, &cacheData)

	if cacheData.Wait == true {

		startWaitTime := public.NowTimeS()
		for {
			ret, err = cm.rdb.Get(cm.ctx, key).Result()

			if err != nil {
				log.DebugError("get value failed", err)
				return
			}

			public.ParserJson(ret, &cacheData)

			if cacheData.D != "" {
				break
			}

			if cacheData.Wait == false {
				break
			}

			public.Sleep(3000)

			nowtime := public.NowTimeS()

			if nowtime-startWaitTime > 600 {
				cm.rdb.Del(cm.ctx, key)
				break
			}
		}
	}

	if cacheData.D == "" {

		var newCacheData StandardcmCache
		newCacheData.Wait = true
		err = cm.rdb.Set(cm.ctx, key, public.BuildJson(newCacheData), time.Duration(maxalivetime)).Err()
		if err != nil {
			log.DebugError("set value failed", err)
			return
		}

		defer func() {
			if r := recover(); r != nil {
				log.DebugError("err:", r)
				cm.rdb.Del(cm.ctx, key)
			}
		}()

		var newData interface{}
		newData, err = newCacheFunc()
		if err != nil {

		}

		ret = public.BuildJson(newData)

		nowtime = public.NowTimeS()

		newCacheData.D = ret
		newCacheData.L = nowtime
		newCacheData.LW = 0
		newCacheData.W = false
		newCacheData.Wait = false

		err = cm.rdb.Set(cm.ctx, key, public.BuildJson(newCacheData), time.Duration(maxalivetime)).Err()
		if err != nil {
			log.DebugError("set value failed", err)
			return
		}

		return

	} else if ((nowtime-cacheData.L >= forceupdatetime) && !cacheData.W) || ((cacheData.LW != 0) && (nowtime-cacheData.LW >= maxworktime)) {

		go func() {
			var newCacheData StandardcmCache
			newCacheData.D = cacheData.D
			newCacheData.L = nowtime
			newCacheData.LW = nowtime
			newCacheData.W = true
			newCacheData.Wait = false

			err := cm.rdb.Set(cm.ctx, key, public.BuildJson(newCacheData), time.Duration(maxalivetime)).Err()
			if err != nil {
				log.DebugError("set value failed", err)
			}

			defer func() {
				if r := recover(); r != nil {
					log.DebugError("err:", r)
					cm.rdb.Del(cm.ctx, key)
				}
			}()

			newData, err := newCacheFunc()
			if err != nil {
				log.DebugError("set value failed", err)
				return
			}

			newDataStr := public.BuildJson(newData)

			nowtime = public.NowTimeS()

			newCacheData.D = newDataStr
			newCacheData.L = nowtime
			newCacheData.LW = 0
			newCacheData.W = false
			newCacheData.Wait = false

			err = cm.rdb.Set(cm.ctx, key, public.BuildJson(newCacheData), time.Duration(maxalivetime)).Err()
			if err != nil {
				log.DebugError("set value failed", err)
				return
			}
		}()

		ret = cacheData.D
		return
	}

	//log.DebugLog("nowtime - cacheData.L: ", (nowtime - cacheData.L), "    forceupdatetime:", forceupdatetime)

	ret = cacheData.D
	return
}

func (cm *CacheManager) DelCache(key string) {
	err := cm.rdb.Del(cm.ctx, key).Err()
	if err != nil {
		log.DebugError("del cache err:", err)
	}
}

func SetCache(key string, value interface{}, configtime ...int64) {
	//configtime[0]	forceupdatetime
	//configtime[1]	maxworktime
	//configtime[2]	maxalivetime

	cacheManager.SetCache(key, value, configtime...)
}

func GetCache(key string, newCacheFunc NewCacheFunc, configtime ...int64) (string, error) {
	//configtime[0]	forceupdatetime
	//configtime[1]	maxworktime
	//configtime[2]	maxalivetime

	return cacheManager.GetCache(key, newCacheFunc, configtime...)
}

func DelCache(key string) {
	cacheManager.DelCache(key)
}

func init() {
	go func() {
		public.Sleep(1000)
		cacheManager.Init(global.AppConfig.Redis.Host, global.AppConfig.Redis.Password, int(global.AppConfig.Redis.DB), global.AppConfig.Redis.EnableTls)
	}()
}
