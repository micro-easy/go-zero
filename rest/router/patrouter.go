package router

import (
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/micro-easy/go-zero/core/search"
	"github.com/micro-easy/go-zero/rest/httpx"
	"github.com/micro-easy/go-zero/rest/internal/context"
)

const (
	allowHeader          = "Allow"
	allowMethodSeparator = ", "
)

var (
	ErrInvalidMethod = errors.New("not a valid http method")
	ErrInvalidPath   = errors.New("path must begin with '/'")
)

type patRouter struct {
	trees      map[string]*search.Tree
	notFound   http.Handler
	notAllowed http.Handler
}

func NewRouter() httpx.Router {
	return &patRouter{
		trees: make(map[string]*search.Tree),
	}
}

func (pr *patRouter) Handle(method, reqPath string, handler http.Handler) error {
	if !validMethod(method) {
		return ErrInvalidMethod
	}

	if len(reqPath) == 0 || reqPath[0] != '/' {
		return ErrInvalidPath
	}

	cleanPath := path.Clean(reqPath)
	if tree, ok := pr.trees[method]; ok {
		return tree.Add(cleanPath, handler)
	} else {
		tree = search.NewTree()
		pr.trees[method] = tree
		return tree.Add(cleanPath, handler)
	}
}

func (pr *patRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqPath := path.Clean(r.URL.Path)
	if tree, ok := pr.trees[r.Method]; ok {
		if result, ok := tree.Search(reqPath); ok {
			if len(result.Params) > 0 {
				r = context.WithPathVars(r, result.Params)
			}
			result.Item.(http.Handler).ServeHTTP(w, r)
			return
		}
	}

	allow, ok := pr.methodNotAllowed(r.Method, reqPath)
	if !ok {
		pr.handleNotFound(w, r)
		return
	}

	if pr.notAllowed != nil {
		pr.notAllowed.ServeHTTP(w, r)
	} else {
		w.Header().Set(allowHeader, allow)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (pr *patRouter) SetNotFoundHandler(handler http.Handler) {
	pr.notFound = handler
}

func (pr *patRouter) SetNotAllowedHandler(handler http.Handler) {
	pr.notAllowed = handler
}

func (pr *patRouter) handleNotFound(w http.ResponseWriter, r *http.Request) {
	if pr.notFound != nil {
		pr.notFound.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func (pr *patRouter) methodNotAllowed(method, path string) (string, bool) {
	var allows []string

	for treeMethod, tree := range pr.trees {
		if treeMethod == method {
			continue
		}

		_, ok := tree.Search(path)
		if ok {
			allows = append(allows, treeMethod)
		}
	}

	if len(allows) > 0 {
		return strings.Join(allows, allowMethodSeparator), true
	} else {
		return "", false
	}
}

func validMethod(method string) bool {
	return method == http.MethodDelete || method == http.MethodGet ||
		method == http.MethodHead || method == http.MethodOptions ||
		method == http.MethodPatch || method == http.MethodPost ||
		method == http.MethodPut
}