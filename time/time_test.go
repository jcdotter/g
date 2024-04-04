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
	"testing"

	"github.com/jcdotter/go/test"
)

var config = &test.Config{
	Trace:   true,
	Detail:  true,
	Require: true,
	Msg:     "%s",
}

func TestTime(t *testing.T) {
	gt := test.New(t, config)

	d := Parse("2006-01-02 15:04:05", "2012-01-31 14:00:00")
	gt.Equal("2012-01-31 14:00:00", d.String(), "time should be equal to '2012-01-31 14:00:00'")
	gt.Equal("2012-01-01", d.YearStart(12).String(), "year start should be equal to '2012-01-01'")
	gt.Equal("2011-02-01", d.YearStart(1).String(), "year start should be equal to '2011-02-01'")
	gt.Equal("2012-01-01", d.QuarterStart(12).String(), "quarter start should be equal to '2012-01-01'")
	gt.Equal("2011-11-01", d.QuarterStart(1).String(), "quarter start should be equal to '2011-11-01'")

}
