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

package messages

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrRateLimited = errors.New("too many submissions")

type quota struct {
	key    string
	limit  int
	window time.Duration
}

func (s *Service) checkQuotas(ctx context.Context, formID, workspaceID uint, ip, email string) error {
	if s.redis == nil {
		return nil
	}

	quotas := []quota{
		{fmt.Sprintf("form:q:ip:%d:%s", formID, ip), s.cfg.PerIPHourly, time.Hour},
		{fmt.Sprintf("form:q:form:%d", formID), s.cfg.PerFormHourly, time.Hour},
		{fmt.Sprintf("form:q:ws:%d", workspaceID), s.cfg.PerWorkspaceDaily, 24 * time.Hour},
	}
	if email != "" {
		quotas = append(quotas, quota{
			fmt.Sprintf("form:q:email:%d:%s", formID, hashKey(email)), s.cfg.PerEmailHourly, time.Hour,
		})
	}

	for _, q := range quotas {
		if q.limit <= 0 || strings.HasSuffix(q.key, ":") {
			continue
		}
		count, err := s.redis.Incr(ctx, q.key).Result()
		if err != nil {
			return ErrRateLimited
		}
		if count == 1 {
			s.redis.Expire(ctx, q.key, q.window)
		}
		if count > int64(q.limit) {
			return ErrRateLimited
		}
	}
	return nil
}

func hashKey(v string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(v)))
	return hex.EncodeToString(sum[:8])
}
