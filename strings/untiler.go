package bookmarks

import (
	"net/http"
	"sort"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/router"

	navpkg "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the bookmarks endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLink("Bookmarks", "/bookmarks", SectionWeight)
	br := r.PathPrefix("/bookmarks").Subrouter()
	r.PathPrefix("/bookmarks").Handler(NewRouter())
	br.NotFoundHandler = http.HandlerFunc(handlers.RenderNotFoundOrLogin)
	br.Use(handlers.IndexMiddleware(bookmarksCustomIndex))
	br.HandleFunc("", BookmarksPage).Methods("GET")
	br.HandleFunc("/mine", MinePage).Methods("GET").MatcherFunc(handlers.RequiresAnAccount())
	br.HandleFunc("/edit", EditPage).Methods("GET").MatcherFunc(handlers.RequiresAnAccount())
	br.HandleFunc("/edit", handlers.TaskHandler(saveTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveTask.Matcher())
	br.HandleFunc("/edit", handlers.TaskHandler(createTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(createTask.Matcher())

}

func NewRouter() *Router {
	return &Router{
		consumeUntilSlash: NewConsumeUntiler("/"),
	}
}

func NewConsumeUntiler(s ...string) ConsumeUntiler {
	matchers := map[int]map[string]struct{}{}
	var sizes []int
	for _, se := range s {
		me, ok := matchers[len(se)]
		if !ok || me == nil {
			me = map[string]struct{}{}
			sizes = append(sizes, len(se))
		}
		me[se] = struct{}{}
		matchers[len(se)] = me
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))
	return ConsumeUntiler{
		matchers: matchers,
		sizes:    sizes,
	}
}

// Register registers the bookmarks router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("bookmarks", nil, RegisterRoutes)
}

type Router struct {
	consumeUntilSlash ConsumeUntiler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Path
	var ok bool
	_, _, p, ok = r.consumeUntilSlash.Consume(p, Inclusive(true))
	if !ok {
		handlers.RenderNotFoundOrLogin(w, req)
		return
	}
}

type ConsumeUntiler struct {
	matchers map[int]map[string]struct{}
	sizes    []int
}

// Consume scans the input string 'from' for any of the configured separators.
// It returns four values:
// 1. matched: The substring before the found separator.
// 2. separator: The separator that was found.
// 3. remaining: The rest of the string. If inclusive is true, this starts after the separator. If false, it starts at the separator.
// 4. found: True if a separator was found, false otherwise.
// If no separator is found, it returns ("", "", from, false).
func (cu ConsumeUntiler) Consume(from string, ops ...any) (string, string, string, bool) {
	inclusive := false
	startOffset := 0
	ignore0PositionMatch := false
	for _, op := range ops {
		switch v := op.(type) {
		case Inclusive:
			inclusive = bool(v)
		case StartOffset:
			startOffset = int(v)
		case Ignore0PositionMatch:
			ignore0PositionMatch = bool(v)
		}
	}
	for i := startOffset; i < len(from); i++ {
		for _, size := range cu.sizes {
			if i+size > len(from) {
				continue
			}
			extract := from[i : i+size]
			if _, ok := cu.matchers[size][extract]; ok {
				if i == 0 && ignore0PositionMatch {
					continue
				}
				matched := from[:i]
				if inclusive {
					return matched + extract, extract, from[i+size:], true
				}
				return matched, extract, from[i:], true
			}
		}
	}
	return "", "", from, false
}

type Inclusive bool
type StartOffset int
type Ignore0PositionMatch bool
