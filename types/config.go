package types

import "github.com/Ptt-official-app/go-pttbbs/bbs"

func config() {
	SERVICE_MODE = ServiceMode(setStringConfig("SERVICE_MODE", string(SERVICE_MODE)))

	HTTP_HOST = setStringConfig("HTTP_HOST", HTTP_HOST)
	URL_PREFIX = setStringConfig("URL_PREFIX", URL_PREFIX)
	BACKEND_PREFIX = setStringConfig("BACKEND_PREFIX", BACKEND_PREFIX)

	PTTSYSOP = bbs.UUserID(setStringConfig("PTTSYSOP", string(PTTSYSOP)))

	STATIC_DIR = setStringConfig("STATIC_DIR", STATIC_DIR)

	ALLOW_ORIGINS = setListStringConfig("ALLOW_ORIGINS", ALLOW_ORIGINS)
	BLOCKED_REFERERS = setListStringConfig("BLOCKED_REFERERS", BLOCKED_REFERERS)
	IS_ALLOW_CROSSDOMAIN = setBoolConfig("IS_ALLOW_CROSSDOMAIN", IS_ALLOW_CROSSDOMAIN)

	COOKIE_DOMAIN = setStringConfig("COOKIE_DOMAIN", COOKIE_DOMAIN)
	TOKEN_COOKIE_SUFFIX = setStringConfig("TOKEN_COOKIE_SUFFIX", TOKEN_COOKIE_SUFFIX)

	CSRF_SECRET = setBytesConfig("CSRF_SECRET", CSRF_SECRET)
	CSRF_TOKEN = setStringConfig("CSRF_TOKEN", CSRF_TOKEN)
	CSRF_TOKEN_TS = setIntConfig("CSRF_TOKEN_TS", CSRF_TOKEN_TS)

	ACCESS_TOKEN = setStringConfig("ACCESS_TOKEN", ACCESS_TOKEN)
	ACCESS_TOKEN_EXPIRE_TS = setIntConfig("ACCESS_TOKEN_EXPIRE_TS", ACCESS_TOKEN_EXPIRE_TS)

	BIG5_TO_UTF8 = setStringConfig("BIG5_TO_UTF8", BIG5_TO_UTF8)
	UTF8_TO_BIG5 = setStringConfig("UTF8_TO_BIG5", UTF8_TO_BIG5)
	TIME_LOCATION = setStringConfig("TIME_LOCATION", TIME_LOCATION)
}
