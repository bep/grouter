// Package grouter implements GopherJS bindings for react-router.
package grouter

import (
	"github.com/bep/gr"
	"github.com/bep/gr/support"
	"github.com/gopherjs/gopherjs/js"
)

// Router represents a router.
type Router struct {
	*js.Object

	root Route

	history History
}

// History for the Router.
type History struct {
	*js.Object
}

// WithHistory is a configuration option to set the history implementation to use.
func WithHistory(history History) func(*Router) error {
	return func(rc *Router) error {
		rc.history = history
		return nil
	}
}

// A Route is defined by a path and one or more components. Components can be
// named. Routes can be nested..
type Route struct {
	path string

	// Either a single component or named components.
	components Components
	component  gr.Component

	children []Route
}

// Components represents named components for a route.
type Components map[string]gr.Component

// NewRoute creates a new Route with the given path and named components and
// the supplied child routes.
func NewRoute(path string, components Components, children ...Route) Route {
	return Route{path: path, components: components, children: children}
}

// NewIndexRoute defines the route used when the user visits "/".
// See https://github.com/reactjs/react-router/blob/master/docs/guides/IndexRoutes.md
func NewIndexRoute(components Components) Route {
	// For now the index route is defined as a route without path.
	// This should be made more robust.
	// TODO(bep)
	return Route{components: components}
}

// New creates a new Router with the given root path and component and options.
func New(path string, c gr.Component, options ...func(*Router) error) *Router {
	root := Route{path: path, component: c}
	router := &Router{root: root}
	router.history = defaultHistory

	for _, opt := range options {
		err := opt(router)
		if err != nil {
			panic(err)
		}
	}

	return router
}

// With creates a new Router with the provided children.
func (r Router) With(routes ...Route) *Router {
	r.Object = nil
	r.root.children = routes
	return &r
}

// Node creates a new React JS component of the Router defintion.
func (r *Router) Node() *js.Object {
	if r.Object == nil {
		r.initObject()
	}
	return r.Object
}

func (r *Router) initObject() {
	// TODO(bep) make annotated struct
	routerProps := make(map[string]interface{})
	routerProps["history"] = r.history

	rootProps := js.Global.Get("Object").New()

	// Root has only one.
	rootProps.Set("component", r.root.component.Node())

	rootProps.Set("path", r.root.path)
	rootProps.Set("key", "root")

	childElements := extractDescendants(r.root.children)
	rootElement := routeFactory.Invoke(rootProps, childElements)
	router := routerFactory.Invoke(routerProps, rootElement)

	if router == nil || router == js.Undefined {
		panic("Failed to create routes")
	}

	r.Object = router
}

func extractDescendants(children []Route) *js.Object {
	childElements := js.Global.Get("Array").New(children)

	for i, c := range children {
		props := js.Global.Get("Object").New()
		props.Set("key", i)
		if c.path != "" {
			props.Set("path", c.path)
		}

		comps := js.Global.Get("Object").New()

		for k, v := range c.components {
			comps.Set(k, v.Node())
		}

		props.Set("components", comps)

		var descendants *js.Object

		if len(c.children) > 0 {
			// Recurse
			descendants = extractDescendants(c.children)
		}

		factory := routeFactory

		if c.path == "" {
			factory = indexRouteFactory
		}

		childElements.SetIndex(i, factory.Invoke(props, descendants))
	}

	return childElements

}

// Link creates a HTML anchor to the given path with the given text.
func Link(to, text string) gr.Modifier {
	p := gr.Props{"to": to, "activeClassName": "active"}
	n := linkFactory.Invoke(p, text)
	return gr.NewPreparedElement(n)
}

// IsActive returns whether the path or location is active.
func IsActive(props gr.Props, pathOrLoc string) bool {
	return getRouterFunc(props, "isActive")(pathOrLoc).Bool()
}

// MarkIfActive marks by the active CSS class if the given path or location
// is active or not. If not active, a modifier that does nothing is returned.
func MarkIfActive(props gr.Props, pathOrLoc string) gr.Modifier {
	if IsActive(props, pathOrLoc) {
		return gr.CSS("active")
	}
	return gr.Discard
}

// WithRouter must be applied to the React component to get hold of the
// router object needed by MarkIfActive and friends.
func WithRouter(o *js.Object) *js.Object {
	return withRouter.Invoke(o)
}

func getRouterFunc(props gr.Props, funcName string) func(...interface{}) *js.Object {
	if r := props.Interface("router"); r != nil {
		router := r.(map[string]interface{})
		if fi, ok := router[funcName]; ok {
			return fi.(func(...interface{}) *js.Object)
		}
		panic(funcName + " not found")

	}

	panic("router not found in props, make sure to decorate your component with WithRouter.")

}

var (
	react             *js.Object
	reactRouter       *js.Object
	routeFactory      *js.Object
	indexRouteFactory *js.Object
	routerFactory     *js.Object
	linkFactory       *js.Object
	defaultHistory    History
	withRouter        *js.Object
)

func init() {

	react := js.Global.Get("React")
	var err error

	if react == js.Undefined {
		// Fallback to Require
		if react, err = support.Require("react"); err != nil {
			panic("Facebook React not found, make sure it is loaded.")
		}
	}

	reactRouter = js.Global.Get("ReactRouter")

	if reactRouter == js.Undefined {
		// Fallback to Require
		if react, err = support.Require("react-router"); err != nil {
			panic("Make sure that react-router is loaded, see https://github.com/reactjs/react-router")
		}
	}

	routeFactory = react.Call("createFactory", reactRouter.Get("Route"))

	if routeFactory == js.Undefined {
		panic("ReactRouter.Route not found.")
	}

	indexRouteFactory = react.Call("createFactory", reactRouter.Get("IndexRoute"))

	if indexRouteFactory == js.Undefined {
		panic("ReactRouter.IndexRoute not found.")
	}

	routerFactory = react.Call("createFactory", reactRouter.Get("Router"))

	if routerFactory == js.Undefined {
		panic("ReactRouter.Router not found.")
	}

	linkFactory = react.Call("createFactory", reactRouter.Get("Link"))

	if linkFactory == js.Undefined {
		panic("ReactRouter.Link not found.")
	}

	withRouter = reactRouter.Get("withRouter")

	if withRouter == js.Undefined {
		panic("ReactRouter.withRouter not found.")
	}

	defaultHistory = History{Object: reactRouter.Get("hashHistory")}

}
