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

package email

import (
	"mime"
	"net/mail"
	"regexp"
	"strings"
	"testing"
)

var encodedWordRe = regexp.MustCompile(`=\?[^?\s]+\?[QqBb]\?[^?\s]*\?=`)

// RFC 2047 ignores the whitespace between encoded-words, but some relays keep
// it, so a subject split mid-word arrives as "subscri ptions". Encoded-words
// must therefore break only at spaces; such a relay may still double one.
func TestBuildMessageSubjectSplitsBetweenWords(t *testing.T) {
	for _, subject := range []string{
		"Your availability alert — new rooms found for 3 subscriptions",
		"Réservation confirmée : votre séjour à l’hôtel Côte d’Azur commence bientôt",
		"Места появились по 3 подпискам",
	} {
		raw := buildMessage("sender@example.com", []string{"rcpt@example.com"},
			subject, "", "body", nil, nil, "", "", false)

		msg, err := mail.ReadMessage(strings.NewReader(string(raw)))
		if err != nil {
			t.Fatalf("built message does not parse: %v", err)
		}
		value := msg.Header.Get("Subject")
		if got, _ := new(mime.WordDecoder).DecodeHeader(value); got != subject {
			t.Errorf("subject %q decodes as %q", subject, got)
		}

		naive := encodedWordRe.ReplaceAllStringFunc(value, func(w string) string {
			s, _ := new(mime.WordDecoder).Decode(w)
			return s
		})
		if strings.Join(strings.Fields(naive), " ") != subject {
			t.Errorf("subject %q arrives as %q\nencoded: %s", subject, naive, value)
		}
	}
}
