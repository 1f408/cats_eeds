package tmplext

import (
	"encoding/base64"

	"github.com/1f408/cats_eeds/internal/perenc"
)

func TmplBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

var TmplUrlPath = perenc.EncodeUrlPath
var TmplUrlFragment = perenc.EncodeUrlFragment
var TmplUriData = perenc.EncodeUriData
var TmplHref = perenc.EncodeHref
var TmplUrl = perenc.EncodeUrl
