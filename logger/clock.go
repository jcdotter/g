// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://github.com/jcdotter/grpg/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"math"
	"sync"
	"time"
)

type clock struct {
	sync.Mutex
	time    time.Time
	fmt     string
	cache   []byte
	nsecPos int
	sCache  clockCache
	nCache  clockCache
}

type clockCache struct {
	exp   int64
	cache []byte
	fmt   string
}

func Clock() *clock {
	return &clock{time: time.Now().UTC()}
}

func (c *clock) Format(f string) *clock {
	c.fmt = f
	c.sCache.fmt, c.nCache.fmt, c.nsecPos = parseFracSec(f)
	c.refreshSec()
	return c
}

func (c *clock) refresh() bool {
	c.Lock()
	defer c.Unlock()
	c.time = time.Now()
	if c.sCache.exp < c.time.Unix() {
		c.refreshSec()
		return true
	} else if c.nCache.exp < int64(c.time.Nanosecond()) {
		c.refreshNsec()
		return true
	}
	return false
}

func (c *clock) refreshSec() {
	c.cache = []byte(c.time.Format(c.fmt))
	e := c.nsecPos + len(c.nCache.fmt)
	// update sec cache
	c.sCache.exp = c.time.Unix() + 1
	c.sCache.cache = append(c.cache[:c.nsecPos], c.cache[e:]...)
	// update nsec cache
	c.nCache.exp = int64(c.time.Nanosecond()) + int64(math.Pow10(10-len(c.nCache.fmt)))
	c.nCache.cache = c.cache[c.nsecPos:e]
}

func (c *clock) refreshNsec() {
	c.nCache.exp = int64(c.time.Nanosecond()) + int64(math.Pow10(10-len(c.nCache.fmt)))
	c.nCache.cache = []byte(c.time.Format(c.nCache.fmt))
	c.cache = append(append(c.sCache.cache[:c.nsecPos], c.nCache.cache...), c.sCache.cache[c.nsecPos:]...)
}

func (c *clock) Unix() int64 {
	return time.Now().UTC().Unix()
}

func (c *clock) UnixMilli() int64 {
	return time.Now().UTC().UnixMilli()
}

func (c *clock) UnixNano() int64 {
	return time.Now().UTC().UnixNano()
}

func (c *clock) Sec() int64 {
	return time.Now().UTC().Unix()
}

func (c *clock) Nanosecond() int64 {
	return int64(time.Now().UTC().Nanosecond())
}

func (c *clock) Millisecond() int64 {
	return int64(time.Now().UTC().Nanosecond() / 1000000)
}

func (c *clock) Time() time.Time {
	return time.Now().UTC()
}

func (c *clock) String() string {
	c.refresh()
	return string(c.cache)
}

func parseFracSec(f string) (fmt, nFmt string, nPos int) {
	var s, e int
	for i := 0; i < len(f); i++ {
		if f[i] == '.' {
			s = i
			c := f[i+1]
			if c == '0' || c == '9' {
				i++
				if f[i+1] == c {
					i++
					for ; i < len(f); i++ {
						if f[i] != c {
							e = i
							break
						}
					}
				}
			}
		}
	}
	if s != 0 && e == 0 {
		e = len(f)
	}
	if s != 0 {
		fmt = f[:s]
		nFmt = f[s:e]
		nPos = s
		if e != len(f) {
			fmt += f[e:]
		}
	}
	return
}
