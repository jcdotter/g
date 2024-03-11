// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License";
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

package ui

import "strings"

// -----------------------------------------------------------------------------
// CSS ELEMENTS
/*
LAYOUT
aspect (auto, square, video, 16/9, 4/3, 21/9, 3/2, 1/1)
container (see breakpoints)
columns (1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, auto)
break-after (auto, avoid, always, all, column, page, left, right)
break-before (auto, avoid, always, all, column, page, left, right)
break-inside (auto, avoid, avoid-page, avoid-column)
display (none, block, inline, inline-block, flex, grid, table, table-cell, table-row, table-column, table-column-group, table-footer-group, table-header-group, table-row-group, flow-root, contents, list-item, inline-list-item, inline-table, inline-flex, inline-grid, run-in, ruby, ruby-base, ruby-text, ruby-base-container, ruby-text-container, contents, none, block, inline, inline-block, flex, grid, table, table-cell, table-row, table-column, table-column-group, table-footer-group, table-header-group, table-row-group, flow-root, contents, list-item, inline-list-item, inline-table, inline-flex, inline-grid, run-in, ruby, ruby-base, ruby-text, ruby-base-container, ruby-text-container, contents)
*/

// -----------------------------------------------------------------------------
// COLOR PALETTE

// Tailwind CSS RGB color palette
var (
	Black   = "0, 0, 0"
	White   = "255, 255, 255"
	Opacity = color{"0.05", "0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7", "0.8", "0.9", "0.95"}
	Slate   = color{"248 250 252", "241 245 249", "226 232 240", "203 213 225", "148 163 184", "100 116 139", "71 85 105", "51 65 85", "30 41 59", "15 23 42", "2 6 23"}
	Gray    = color{"249 250 251", "243 244 246", "229 231 235", "209 213 219", "156 163 175", "107 114 128", "75 85 99", "55 65 81", "31 41 55", "17 24 39", "3 7 18"}
	Zinc    = color{"250 250 250", "244 244 245", "228 228 231", "212 212 216", "161 161 170", "113 113 122", "82 82 91", "63 63 70", "39 39 42", "24 24 27", "9 9 11"}
	Neutral = color{"250 250 250", "245 245 245", "229 229 229", "212 212 212", "163 163 163", "115 115 115", "82 82 82", "64 64 64", "38 38 38", "23 23 23", "10 10 10"}
	Stone   = color{"250 250 249", "245 245 244", "231 229 228", "214 211 209", "168 162 158", "120 113 108", "87 83 78", "68 64 60", "41 37 36", "28 25 23", "12 10 9"}
	Red     = color{"254 242 242", "254 226 226", "254 202 202", "252 165 165", "248 113 113", "239 68 68", "220 38 38", "185 28 28", "153 27 27", "127 29 29", "69 10 10"}
	Orange  = color{"255 247 237", "255 237 213", "254 215 170", "253 186 116", "251 146 60", "249 115 22", "234 88 12", "194 65 12", "154 52 18", "124 45 18", "67 20 7"}
	Amber   = color{"255 251 235", "254 243 199", "253 230 138", "252 211 77", "251 191 36", "245 158 11", "217 119 6", "180 83 9", "146 64 14", "120 53 15", "69 26 3"}
	Yellow  = color{"254 252 232", "254 249 195", "254 240 138", "253 224 71", "250 204 21", "234 179 8", "202 138 4", "161 98 7", "133 77 14", "113 63 18", "66 32 6"}
	Lime    = color{"247 254 231", "236 252 203", "217 249 157", "190 242 100", "163 230 53", "132 204 22", "101 163 13", "77 124 15", "63 98 18", "54 83 20", "26 46 5"}
	Green   = color{"240 253 244", "220 252 231", "187 247 208", "134 239 172", "74 222 128", "34 197 94", "22 163 74", "21 128 61", "22 101 52", "20 83 45", "5 46 22"}
	Emerald = color{"236 253 245", "209 250 229", "167 243 208", "110 231 183", "52 211 153", "16 185 129", "5 150 105", "4 120 87", "6 95 70", "6 78 59", "2 44 34"}
	Teal    = color{"240 253 250", "204 251 241", "153 246 228", "94 234 212", "45 212 191", "20 184 166", "13 148 136", "15 118 110", "17 94 89", "19 78 74", "4 47 46"}
	Cyan    = color{"236 254 255", "207 250 254", "165 243 252", "103 232 249", "34 211 238", "6 182 212", "8 145 178", "14 116 144", "21 94 117", "22 78 99", "8 51 68"}
	Sky     = color{"240 249 255", "224 242 254", "186 230 253", "125 211 252", "56 189 248", "14 165 233", "2 132 199", "3 105 161", "7 89 133", "12 74 110", "8 47 73"}
	Blue    = color{"239 246 255", "219 234 254", "191 219 254", "147 197 253", "96 165 250", "59 130 246", "37 99 235", "29 78 216", "30 64 175", "30 58 138", "23 37 84"}
	Indigo  = color{"238 242 255", "224 231 255", "199 210 254", "165 180 252", "129 140 248", "99 102 241", "79 70 229", "67 56 202", "55 48 163", "49 46 129", "30 27 75"}
	Violet  = color{"245 243 255", "237 233 254", "221 214 254", "196 181 253", "167 139 250", "139 92 246", "124 58 237", "109 40 217", "91 33 182", "76 29 149", "46 16 101"}
	Purple  = color{"250 245 255", "243 232 255", "233 213 255", "216 180 254", "192 132 252", "168 85 247", "147 51 234", "126 34 206", "107 33 168", "88 28 135", "59 7 100"}
	Fuchsia = color{"253 244 255", "250 232 255", "245 208 254", "240 171 252", "232 121 249", "217 70 239", "192 38 211", "162 28 175", "134 25 143", "112 26 117", "74 4 78"}
	Pink    = color{"253 242 248", "252 231 243", "251 207 232", "249 168 212", "244 114 182", "236 72 153", "219 39 119", "190 24 93", "157 23 77", "131 24 67", "80 7 36"}
	Rose    = color{"255 241 242", "255 228 230", "254 205 211", "253 164 175", "251 113 133", "244 63 94", "225 29 72", "190 18 60", "159 18 57", "136 19 55", "76 5 25"}
)

// color is a struct that holds the RGB values of a color
type color struct{ s50, s100, s200, s300, s400, s500, s600, s700, s800, s900, s950 string }

func (c color) rbg(s int) string {
	switch s {
	case 50:
		return c.s50
	case 100:
		return c.s100
	case 200:
		return c.s200
	case 300:
		return c.s300
	case 400:
		return c.s400
	case 500:
		return c.s500
	case 600:
		return c.s600
	case 700:
		return c.s700
	case 800:
		return c.s800
	case 900:
		return c.s900
	default:
		return c.s950
	}
}

// RGB returns a css string with the RGB value
// of the color of the given shade.
func (c color) RGB(shade int) string {
	b := make([]byte, 0, 24)
	b = append(b, "rgb("...)
	b = append(b, c.rbg(shade)...)
	b = append(b, ");"...)
	return string(b)
}

// RGBA returns a css string with the RGBA value
// of the color of the given shade and opacity.
func (c color) RGBA(shade int, opacity int) string {
	b := make([]byte, 0, 24)
	b = append(b, "rgba("...)
	for _, s := range strings.Split(c.rbg(shade), " ") {
		b = append(b, s...)
		b = append(b, ',')
	}
	b = append(b, Opacity.rbg(opacity)...)
	b = append(b, ");"...)
	return string(b)
}
