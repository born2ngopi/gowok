package gowok

import (
	"net/http"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/ngamux/ngamux"
)

type HttpDocs struct {
	swagger *spec.Swagger
}

type HttpDocsItem struct {
	*spec.PathItemProps
	method string
	path   string
}

func NewHttpDocs(title, version string) *HttpDocs {
	swagger := spec.Swagger{
		VendorExtensible: spec.VendorExtensible{},
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Schemes: []string{"http", "ws"},
			Host:    "localhost:8080",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Version: version,
					Title:   title,
				},
			},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{},
			},
		},
	}

	return &HttpDocs{&swagger}
}

func (docs *HttpDocs) NewItem(method, path string, operation *spec.Operation) *HttpDocsItem {
	item := spec.PathItemProps{}
	if itemFound, ok := docs.swagger.Paths.Paths[path]; ok {
		item = itemFound.PathItemProps
	}

	if operation == nil {
		operation = spec.NewOperation(strings.Join([]string{method, path}, "-"))
	}
	if operation.Description == "" {
		operation.Description = operation.ID
	}

	switch method {
	case http.MethodGet:
		item.Get = operation
	case http.MethodPost:
		item.Post = operation
	case http.MethodPut:
		item.Put = operation
	case http.MethodHead:
		item.Head = operation
	case http.MethodPatch:
		item.Patch = operation
	case http.MethodDelete:
		item.Delete = operation
	case http.MethodOptions:
		item.Options = operation
	}
	docs.swagger.Paths.Paths[path] = spec.PathItem{
		PathItemProps: item,
	}

	return &HttpDocsItem{&item, method, path}
}

func (docs HttpDocs) ServeHTTP(rw http.ResponseWriter, r *http.Request) error {
	return ngamux.Res(rw).JSON(docs.swagger)
}

func (docs *HttpDocs) Get(path string, operation *spec.Operation) *HttpDocsItem {
	return docs.NewItem(http.MethodGet, path, operation)
}
func (docs *HttpDocs) Post(path string, operation *spec.Operation) *HttpDocsItem {
	return docs.NewItem(http.MethodPost, path, operation)
}
func (docs *HttpDocs) Put(path string, operation *spec.Operation) *HttpDocsItem {
	return docs.NewItem(http.MethodPut, path, operation)
}
func (docs *HttpDocs) Head(path string, operation *spec.Operation) *HttpDocsItem {
	return docs.NewItem(http.MethodHead, path, operation)
}
func (docs *HttpDocs) Patch(path string, operation *spec.Operation) *HttpDocsItem {
	return docs.NewItem(http.MethodPatch, path, operation)
}
func (docs *HttpDocs) Delete(path string, operation *spec.Operation) *HttpDocsItem {
	return docs.NewItem(http.MethodDelete, path, operation)
}
func (docs *HttpDocs) Options(path string, operation *spec.Operation) *HttpDocsItem {
	return docs.NewItem(http.MethodOptions, path, operation)
}

func (item *HttpDocsItem) Handle(mux *ngamux.HttpServeMux, handlers ngamux.Handler) {
	mux.HandlerFunc(
		item.method,
		item.path,
		handlers,
	)
}