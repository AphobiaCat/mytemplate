package timermanager

import (
	"mytemplate/pkg/log"
	"time"
)

type TimerCallback func()

const maxtimercons int = 10

var timermanager TimerManager

type TimerManager struct {
	maxtimerindex   int
	nowtimerindex   int
	callbacks       [maxtimercons]TimerCallback
	callbackmaxtime [maxtimercons]uint32
	callbacknowtime [maxtimercons]uint32

	initonce     bool
	ticker       *time.Ticker
	nowtime      uint32
	timerongoing bool
}

func (tm *TimerManager) Init() {

	if tm.initonce == false {
		tm.maxtimerindex = maxtimercons
		tm.nowtimerindex = 0
	}
	tm.initonce = true

	tm.ticker = time.NewTicker(1 * time.Second)
	tm.nowtime = 0

	tm.timerongoing = true

	go func() {
		defer tm.ticker.Stop()
		for tm.timerongoing {
			select {
			case _ = <-tm.ticker.C:
				tm.nowtime++
			}

			for i := 0; i < tm.nowtimerindex; i++ {
				tm.callbacknowtime[i]++
				if tm.callbacknowtime[i] >= tm.callbackmaxtime[i] {
					tm.callbacknowtime[i] = 0
					go tm.callbacks[i]()
				}
			}
		}
	}()
}

func Stop() {
	timermanager.timerongoing = false
	timermanager.nowtimerindex = 0
}

func RegTimer(calltimes uint32, callback TimerCallback) int {
	if timermanager.nowtimerindex >= timermanager.maxtimerindex {
		log.DebugLog("should extented timer")
		return -1
	}

	timermanager.callbacknowtime[timermanager.nowtimerindex] = 0
	timermanager.callbackmaxtime[timermanager.nowtimerindex] = calltimes
	timermanager.callbacks[timermanager.nowtimerindex] = callback

	timermanager.nowtimerindex++

	return timermanager.nowtimerindex - 1
}

func ChangeTimerSet(timerid int, calltimes uint32) bool {
	if timerid >= timermanager.maxtimerindex {
		log.DebugLog("error timer id")
		return false
	}

	timermanager.callbacknowtime[timerid] = 0
	timermanager.callbackmaxtime[timerid] = calltimes

	return true
}

func init() {
	timermanager.Init()
}
