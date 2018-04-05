package routing

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type key int

const (
	requestKey key = 1
	paramsKey  key = 2
	langKey    key = 3
	groupKey   key = 4
)

func Param(r *http.Request, name string) string {
	return r.Context().Value(paramsKey).(httprouter.Params).ByName(name)
}

func RequestFromContext(ctx context.Context) *http.Request {
	return ctx.Value(requestKey).(*http.Request)
}

func GroupFromContext(ctx context.Context) Group {
	return ctx.Value(groupKey).(Group)
}

func LangFromContext(ctx context.Context) string {
	return ctx.Value(langKey).(string)
}
