// Package grouter implements GopherJS bindings for react-router.

package grouter

import (
	"github.com/bep/gr"
	"github.com/gopherjs/gopherjs/js"
)

type Router struct {
	element *gr.Element
	root    Route
}

type Route struct {
	path string

	// Either a single component or named components.
	components Components
	component  gr.Component

	children []Route
}

type Components map[string]gr.Component

func NewRoute(path string, components Components, children ...Route) Route {
	return Route{path: path, components: components, children: children}
}

func New(path string, c gr.Component, children ...Route) *Router {
	root := Route{path: path, component: c, children: children}
	return &Router{root: root}
}

func (r *Router) Node() *js.Object {

	routerProps := make(map[string]interface{})
	routerProps["history"] = hashHistory

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

	return router
}

func extractDescendants(children []Route) *js.Object {
	childElements := js.Global.Get("Array").New(children)

	for i, c := range children {
		props := js.Global.Get("Object").New()
		props.Set("key", i)
		props.Set("path", c.path)

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
		childElements.SetIndex(i, routeFactory.Invoke(props, descendants))
	}

	return childElements

}

func Link(to, text string) gr.Modifier {
	p := gr.Props{"to": to, "activeClassName": "active"}
	n := linkFactory.Invoke(p, text)
	return gr.NewPreparedElement(n)
}

func IsActive(props gr.Props, pathOrLoc string) bool {
	return getRouterFunc(props, "isActive")(pathOrLoc).Bool()
}

func MarkIfActive(props gr.Props, pathOrLoc string) gr.Modifier {
	var m gr.Modifier = gr.Discard
	if IsActive(props, pathOrLoc) {
		return gr.CSS("active")
	}
	return m
}

func WithRouter(o *js.Object) *js.Object {
	return withRouter.Invoke(o)
}

func getRouterFunc(props gr.Props, funcName string) func(...interface{}) *js.Object {
	var f func(...interface{}) *js.Object
	if r, ok := props["router"]; ok {
		router := r.(map[string]interface{})
		if fi, ok := router[funcName]; ok {
			f = fi.(func(...interface{}) *js.Object)
		} else {
			panic(funcName + " not found")
		}
	} else {
		panic("router not found in props, make sure to decorate your component with WithRouter.")
	}
	return f
}

var (
	react         *js.Object
	reactRouter   *js.Object
	routeFactory  *js.Object
	routerFactory *js.Object
	linkFactory   *js.Object
	hashHistory   *js.Object
	withRouter    *js.Object
)

func init() {

	react := js.Global.Get("React")

	if react == js.Undefined {
		panic("Facebook React not found, make sure it is loaded.")
	}

	reactRouter = js.Global.Get("ReactRouter")

	if reactRouter == js.Undefined {
		panic("Make sure that react-router is loaded, see https://github.com/reactjs/react-router")
	}

	routeFactory = react.Call("createFactory", reactRouter.Get("Route"))

	if routeFactory == js.Undefined {
		panic("ReactRouter.Route not found.")
	}

	routerFactory = react.Call("createFactory", reactRouter.Get("Router"))

	if routerFactory == js.Undefined {
		panic("ReactRouter.Router not found.")
	}

	linkFactory = react.Call("createFactory", reactRouter.Get("Link"))

	if linkFactory == js.Undefined {
		panic("ReactRouter.Link not found.")
	}

	hashHistory = reactRouter.Get("hashHistory")

	if hashHistory == js.Undefined {
		panic("ReactRouter.hashHistory not found.")
	}

	withRouter = reactRouter.Get("withRouter")

	if withRouter == js.Undefined {
		panic("ReactRouter.withRouter not found.")
	}

}
