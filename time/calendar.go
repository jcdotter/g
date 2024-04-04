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
	"time"
)

type Holiday struct {
	time.Time
	// Name is the common name of the holiday
	Name string
	// Date returns the date of the holiday for year 'y'
	Date func(y int) *Time
}

type Holidays []Holiday

var UsHolidays Holidays

func init() {
	UsHolidays = make(Holidays, 12)
	UsHolidays = append(UsHolidays, Holiday{Name: "New Years Day", Date: NewYears})
	UsHolidays = append(UsHolidays, Holiday{Name: "Martin Luther King Day", Date: MlkDay})
	UsHolidays = append(UsHolidays, Holiday{Name: "Inauguration Day", Date: InagurationDay})
	UsHolidays = append(UsHolidays, Holiday{Name: "Presidents Day", Date: PresidentsDay})
	UsHolidays = append(UsHolidays, Holiday{Name: "Memorial Day", Date: MemorialDay})
	UsHolidays = append(UsHolidays, Holiday{Name: "Juneteenth", Date: NationalIndependenceDay})
	UsHolidays = append(UsHolidays, Holiday{Name: "Independence Day", Date: IndependenceDay})
	UsHolidays = append(UsHolidays, Holiday{Name: "Labor Day", Date: LaborDay})
	UsHolidays = append(UsHolidays, Holiday{Name: "Columbus Day", Date: ColumbusDay})
	UsHolidays = append(UsHolidays, Holiday{Name: "Veterans Day", Date: VeteransDay})
	UsHolidays = append(UsHolidays, Holiday{Name: "Thanksgiving", Date: Thanksgiving})
	UsHolidays = append(UsHolidays, Holiday{Name: "Christmas", Date: Christmas})
}

func (h *Holidays) IsHoliday(t *Time) bool {
	y, m, d := t.Year(), t.Month(), t.Day()
	for _, i := range UsHolidays {
		h := i.Date(y)
		if m == h.Month() && d == h.Day() {
			return true
		}
	}
	return false
}

// Instance returns the date of the 'i' instance of weekday 'wd'
// in month 'm' of year 'y'; if i < 0 returns the last instance, and
// panics if 'i' is 0 or exceeds the number of instances
func Instance(i, wd, m, y int) *Time {
	s := Date(y, m, 1)
	e := s.MonthEnd().Add(-24*time.Hour + time.Nanosecond)
	f := int(s.Weekday())
	l := int(e.Weekday())
	o := 0
	if i < 0 {
		if wd > l {
			o = 7
		}
		return &Time{
			Time: e.AddDate(0, 0, int(wd-l)-o),
			fmt:  s.fmt,
		}
	}
	if wd >= f {
		o = 7
	}
	r := s.AddDate(0, 0, i*7+int(wd-f)-o)
	if r.After(e) || r.Before(s.Time) {
		panic("instance must be greater than 0 and not exceed instances in month")
	}
	return r
}

// NewYears returns the observed date for new years day of for year 'y'
func NewYears(y int) *Time {
	return HolidayObserved(Date(y, 1, 1))
}

// MlkDay returns the date of Martin Luther King Jr Day for year 'y'
func MlkDay(y int) *Time {
	return Instance(3, 1, 1, y)
}

// InagurationDay returns the date of the presidential inaguration for year 'y'
func InagurationDay(y int) *Time {
	y -= (y - 1) % 4
	return HolidayObserved(Date(y, 1, 20))
}

// PresidentsDay returns the date of President's Day (or Washington's Birthday) for year 'y'
func PresidentsDay(y int) *Time {
	return Instance(3, 1, 2, y)
}

// GoodFriday returns the date of good friday for year 'y'
func GoodFriday(y int) *Time {
	return Easter(y).AddDate(0, 0, -2)
}

// Easter returns the date of easter for year 'y'
func Easter(y int) *Time {
	var yr, c, n, k, i, j, l, m, d float64
	yr = float64(y)
	c = math.Floor(yr / 100)
	n = yr - 19*math.Floor(yr/19)
	k = math.Floor((c - 17) / 25)
	i = c - math.Floor(c/4) - math.Floor((c-k)/3) + 19*n + 15
	i = i - 30*math.Floor(i/30)
	i = i - math.Floor(i/28)*(1-math.Floor(i/28)*math.Floor(29/(i+1))*math.Floor((21-n)/11))
	j = yr + math.Floor(yr/4) + i + 2 - c + math.Floor(c/4)
	j = j - 7*math.Floor(j/7)
	l = i - j
	m = 3 + math.Floor((l+40)/44)
	d = l + 28 - 31*math.Floor(m/4)
	return Date(y, int(m), int(d))
}

// MemorialDay returns the date of Memorial Day for year 'y'
func MemorialDay(y int) *Time {
	return Instance(-1, 1, 5, y)
}

// NationalIndependenceDay returns the observed date for
// Junteenth National Independence Day for year 'y'
func NationalIndependenceDay(y int) *Time {
	return HolidayObserved(Date(y, 6, 19))
}

// IndependenceDay returns the observed date for US Independence Day for year 'y'
func IndependenceDay(y int) *Time {
	return HolidayObserved(Date(y, 7, 4))
}

// LaborDay returns the date of Labor Day for year 'y'
func LaborDay(y int) *Time {
	return Instance(1, 1, 9, y)
}

// ColumbusDay returns the date of Columbus Day for year 'y'
func ColumbusDay(y int) *Time {
	return Instance(2, 1, 10, y)
}

// VeteransDay returns the observed date for Veterans Day for year 'y'
func VeteransDay(y int) *Time {
	return HolidayObserved(Date(y, 11, 11))
}

// Thanksgiving returns the date of Thanksgiving Day for year 'y'
func Thanksgiving(y int) *Time {
	return Instance(4, 4, 11, y)
}

// Christmas returns the observed date for Christmas Day for year 'y'
func Christmas(y int) *Time {
	return HolidayObserved(Date(y, 12, 25))
}

// HolidayObserved returns the date holiday 'h' is observed,
// Friday if on Saturday and Monday if on Sunday
func HolidayObserved(h *Time) *Time {
	if h.Weekday() == 6 {
		h = h.AddDate(0, 0, -1)
	} else if h.Weekday() == 0 {
		h = h.AddDate(0, 0, 1)
	}
	return h
}
