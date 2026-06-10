package server

import (
	httpErrors "chronoFlow-exec/internal/errors"
	nethttp "net/http"
)

func errorEncoder(w nethttp.ResponseWriter, r *nethttp.Request, err error) {
	httpErrors.EncodeHTTPError(w, r, err)
}
