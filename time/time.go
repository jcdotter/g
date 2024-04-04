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

import "time"

const (
	ISO8601N   = `2006-01-02 15:04:05.000000000`
	ISO8601    = `2006-01-02 15:04:05.000`
	SqlDate    = `2006-01-02T15:04:05Z`
	TimeFormat = `2006-01-02 15:04:05`
	DateFormat = `2006-01-02`
)

type Time struct{ time.Time }
