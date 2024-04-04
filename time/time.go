// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://github.com/jcdotter/go/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package time

import (
	"math"
	"sync"
	"time"
)

// ----------------------------------------------------------------------------
// TIME FORMATS

const (
	ISO8601N   = `2006-01-02 15:04:05.000000000`
	ISO8601    = `2006-01-02 15:04:05.000`
	SqlDate    = `2006-01-02T15:04:05Z`
	TimeFormat = `2006-01-02 15:04:05`
	DateFormat = `2006-01-02`
)

// ----------------------------------------------------------------------------
// TIME

// Time is a formattable time object
// that caches the formatted time string
// and updates it only when necessary
type Time struct {
	sync.Mutex
	time.Time
	fmt     string
	cache   []byte
	nsecPos int
	sCache  TimeCache
	nCache  TimeCache
}

type TimeCache struct {
	exp   int64
	cache []byte
	fmt   string
}

// Parse returns a new Time
func Parse(f string, t string) *Time {
	tm, _ := time.Parse(f, t)
	return &Time{Time: tm, fmt: f}
}

// Now returns a new Time
func Now() *Time {
	return &Time{Time: time.Now().UTC(), fmt: TimeFormat}
}

// Date returns a new created time using the
// provided year, month, and day
func Date(y, m, d int) *Time {
	return &Time{Time: time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC), fmt: DateFormat}
}

// Format resets the time format
func (t *Time) Format(f string) *Time {
	t.fmt = f
	t.format()
	t.recache()
	return t
}

// Now updates the time to the current time
func (t *Time) Update() *Time {
	t.Refresh()
	return t
}

// Refresh updates the time cache
// to the current time
func (t *Time) Refresh() bool {
	t.Lock()
	defer t.Unlock()
	t.Time = time.Now()
	return t.Cache()
}

// Unix returns the time in seconds
// since the UTC Unix epoch (Jan 1, 1970)
func (t *Time) Unix() int64 {
	return t.Time.UTC().Unix()
}

// UnixMilli returns the time in milliseconds
// since the UTC Unix epoch (Jan 1, 1970)
func (t *Time) UnixMilli() int64 {
	return t.Time.UTC().UnixMilli()
}

// UnixNano returns the time in nanoseconds
// since the UTC Unix epoch (Jan 1, 1970)
func (t *Time) UnixNano() int64 {
	return t.Time.UTC().UnixNano()
}

// String returns the formatted time string
func (t *Time) String() string {
	t.Cache()
	return string(t.cache)
}

// Bytes returns the formatted time string
func (t *Time) Bytes() []byte {
	t.Cache()
	return t.cache
}

// DaysSince returns the number of Days since lt until time t
func (t *Time) DaysSince(lt *Time) int {
	return int(t.Sub(lt.Time).Hours() / 24)
}

// MonthsSince returns the number of full Months since lt until time t
func (t *Time) MonthsSince(lt *Time) int {
	pm := int(float64(lt.Day()) / math.Min(float64(t.Day()), float64(lt.DaysInMonth())))
	return (t.Year() - lt.Year()) + int(t.Month()-lt.Month()) - 1 + pm
}

// YearsSince returns the number of full Years since lt until time t
func (t *Time) YearsSince(lt *Time) int {
	return t.MonthsSince(lt) / 12
}

// DaysInMonth returns the number of calendar days in month of TIME 't'
func (t *Time) DaysInMonth() int {
	return Date(t.Year(), int(t.Month()), 0).AddDate(0, 1, 0).Day()
}

// AddDate returns the time with the added
// year, month, and day as a new Time instance
func (t *Time) AddDate(y, m, d int) *Time {
	return &Time{Time: t.Time.AddDate(y, m, d), fmt: t.fmt}
}

// MonthStart returns the first date of the month for time 't'
func (t *Time) MonthStart() *Time {
	return Date(t.Year(), int(t.Month()), 1)
}

// MonthEnd returns the last nanosecond of the month for time 't'
func (t *Time) MonthEnd() *Time {
	return &Time{
		Time: time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0).Add(-1 * time.Nanosecond),
		fmt:  t.fmt,
	}
}

// QuarterStart returns the first date of the quarter
// for time 't' with year ending in month 'ye'
func (t *Time) QuarterStart(ye int) *Time {
	// pickup here
	ye = ye % 3
	ye = (3-((int(t.Month())-(ye%3))%3))%3 - 2
	return t.AddDate(0, ye, 0).MonthStart()
}

// QuarterEnd returns the last nanosecond of the quarter
// for time 't' with year ending in month 'ye'
func (t *Time) QuarterEnd(ye int) *Time {
	ye = ye % 3
	ye = (3 - ((int(t.Month()) - ye) % 3)) % 3
	return t.AddDate(0, ye, 0).MonthEnd()
}

// YearStart returns the first date of the year
// for time 't' with year ending in month 'ye'
func (t *Time) YearStart(ye int) *Time {
	var y int
	if int(t.Month()) < ye+1 {
		y = 1
	}
	return Date(t.Year()-y, ye+1, 1)
}

// YearEnd returns the last nanosecond of the year
// for time 't' with year ending in month 'ye'
func (t *Time) YearEnd(ye int) *Time {
	return &Time{Time: t.YearStart(ye).Time.AddDate(0, 12, 0).Add(-1 * time.Nanosecond), fmt: t.fmt}
}

func (t *Time) IsHoliday() (Holiday, bool) {
	y, m, d := t.Year(), t.Month(), t.Day()
	for _, i := range GetUsHolidays().List {
		h := i.Date(y)
		if m == h.Month() && d == h.Day() {
			return i, true
		}
	}
	return Holiday{}, false
}

// ----------------------------------------------------------------------------
// HELPS

// format sets the time cache formats
func (t *Time) format() *Time {
	t.sCache.fmt, t.nCache.fmt, t.nsecPos = parseFracSec(t.fmt)
	return t
}

func (t *Time) Cache() bool {
	if t.sCache.exp < t.Time.Unix() {
		t.recache()
		return true
	} else if t.nCache.exp < int64(t.Time.Nanosecond()) {
		t.recacheN()
		return true
	}
	return false
}

func (t *Time) recache() {
	t.cache = []byte(t.Time.Format(t.fmt))
	e := t.nsecPos + len(t.nCache.fmt)
	// update sec cache
	t.sCache.exp = t.Time.Unix() + 1
	t.sCache.cache = append(t.cache[:t.nsecPos], t.cache[e:]...)
	// update nsec cache
	t.nCache.exp = int64(t.Time.Nanosecond()) + int64(math.Pow10(10-len(t.nCache.fmt)))
	t.nCache.cache = t.cache[t.nsecPos:e]
}

func (t *Time) recacheN() {
	t.nCache.exp = int64(t.Time.Nanosecond()) + int64(math.Pow10(10-len(t.nCache.fmt)))
	t.nCache.cache = []byte(t.Time.Format(t.nCache.fmt))
	t.cache = append(append(t.sCache.cache[:t.nsecPos], t.nCache.cache...), t.sCache.cache[t.nsecPos:]...)
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
