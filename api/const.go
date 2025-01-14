package api

import (
	"time"

	"github.com/Ptt-official-app/go-openbbsmiddleware/types"
)

const (
	DEFAULT_MAX_LIST   = 200
	DEFAULT_DESCENDING = true
	DEFAULT_ASCENDING  = true

	ARTICLE_LOCK_TS          = 10
	ARTICLE_LOCK_TS_DURATION = time.Duration(ARTICLE_LOCK_TS) * time.Second
	ARTICLE_LOCK_NANO_TS     = types.NanoTS(ARTICLE_LOCK_TS) * types.TS_TO_NANO_TS

	HTML_CACHE_CONTROL_TS = 3600
	JS_CACHE_CONTROL_TS   = 86400 * 365
)

var (
	GET_ARTICLE_CONTENT_INFO_TOO_SOON_NANO_TS = ARTICLE_LOCK_NANO_TS + types.NanoTS(1)*types.TS_TO_NANO_TS //10 + 1 seconds.

	MIME_TYPE_MAP = map[string]string{
		".js":   "text/javascript",
		".html": "text/html",
		".map":  "application/json",
	}
)
