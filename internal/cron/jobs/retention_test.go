// Copyright (c) 2026 Jonas Kaninda.
//
// This file is part of Posta.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package jobs

import "testing"

// capDays encodes "content never outlives the record": a positive window is
// clamped to the record window; 0 ("forever") is never capped or used as a cap.
func TestCapDays(t *testing.T) {
	cases := []struct {
		v, limit, want int
	}{
		{v: 30, limit: 180, want: 30},   // below the cap
		{v: 180, limit: 180, want: 180}, // at the cap
		{v: 999, limit: 180, want: 180}, // above the cap → clamped
		{v: 0, limit: 180, want: 0},     // v = forever → not capped
		{v: 30, limit: 0, want: 30},     // record = forever → no cap
		{v: 0, limit: 0, want: 0},       // both forever
	}
	for _, c := range cases {
		if got := capDays(c.v, c.limit); got != c.want {
			t.Errorf("capDays(%d, %d) = %d, want %d", c.v, c.limit, got, c.want)
		}
	}
}

// minDays encodes "raw .eml never outlives either content window": the shorter
// finite window wins, and 0 ("forever") means no bound on that side.
func TestMinDays(t *testing.T) {
	cases := []struct {
		a, b, want int
	}{
		{a: 30, b: 90, want: 30}, // shorter first
		{a: 90, b: 30, want: 30}, // shorter second
		{a: 30, b: 30, want: 30}, // equal
		{a: 0, b: 30, want: 30},  // a forever → b bounds
		{a: 30, b: 0, want: 30},  // b forever → a bounds
		{a: 0, b: 0, want: 0},    // both forever
	}
	for _, c := range cases {
		if got := minDays(c.a, c.b); got != c.want {
			t.Errorf("minDays(%d, %d) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
