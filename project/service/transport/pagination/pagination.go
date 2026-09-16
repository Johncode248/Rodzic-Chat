// Package pagination centralizes the ?limit=/?offset= parsing every list
// endpoint in this codebase needs -- previously each handler that supported
// pagination (chats.SelectChannels, caregivers.List) duplicated this same
// parse-clamp-default logic inline. A shared helper means every endpoint
// gets identical, correct behavior (silently ignoring a non-numeric or
// negative value rather than erroring, clamping to a sane max) instead of
// each one being a slightly different hand-copy of the pattern.
package pagination

import (
	"net/http"
	"strconv"
)

// DefaultMaxLimit bounds ?limit= for endpoints that don't need a smaller
// cap of their own -- without a bound, ?limit=999999999 would force the
// backend to load (and JSON-encode) an entire table in one request.
const DefaultMaxLimit = 200

// Parse reads ?limit=/?offset= from the request, clamping limit to
// [1, maxLimit] when present and non-positive/invalid values (including
// omitted params) are simply ignored -- callers apply the returned limit
// only when it's > 0, and offset only when it's > 0, matching the
// zero-value-means-unset convention databasepg.Query already uses.
func Parse(r *http.Request, maxLimit int) (limit int, offset int) {
	if maxLimit <= 0 {
		maxLimit = DefaultMaxLimit
	}

	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		if l > maxLimit {
			l = maxLimit
		}
		limit = l
	}

	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && o > 0 {
		offset = o
	}

	return limit, offset
}
