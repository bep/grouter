package tests

import (
	"testing"

	"github.com/bep/gr"
	"github.com/bep/gr/el"
	"github.com/bep/gr/tests/grt"
	"github.com/bep/grouter"
	"github.com/gopherjs/gopherjs/js"
)

func TestRouter(t *testing.T) {

	var (
		c1    = gr.New(&testComp{name: "c1"})
		c2    = gr.New(&testComp{name: "c2"})
		c2Sub = gr.New(&testComp{name: "c2_sub"})
	)

	// WithRouter makes this.props.router happen.
	appComponent := gr.New(new(testApp), gr.Apply(grouter.WithRouter))

	routerConfig := grouter.New("/", appComponent, grouter.WithHistory(forPath("/c2")))

	router := routerConfig.With(
		grouter.NewIndexRoute(grouter.Components{"main": c1}),
		grouter.NewRoute("c1", grouter.Components{"main": c1}),
		grouter.NewRoute("c2", grouter.Components{"main": c2, "sub": c2Sub}),
	)

	rc := gr.NewSimpleComponent(router)
	elem := rc.CreateElement(nil)

	r := grt.ShallowRender(elem)
	routerContext := r.Dive("Router")

	loc := routerContext.Props["location"].(map[string]interface{})

	// TODO(bep) ... can do better.
	grt.Equal(t, loc["pathname"], "/c2")
}

type testApp struct {
	*gr.This
}

// Implements the Renderer interface.
func (a testApp) Render(this *gr.This) gr.Component {
	return el.Div(
		el.UnorderedList(
			a.createLinkListItem("/c1", "C #1"),
			a.createLinkListItem("/c2", "C #2"),
		),
		// Receives the component in this.props.<name>
		// If none found, a no-op is returned.
		a.Component("main"),
		a.Component("sub"),
	)
}

func (a testApp) createLinkListItem(path, title string) gr.Modifier {
	return el.ListItem(
		grouter.MarkIfActive(a.Props(), path),
		grouter.Link(path, title))
}

type testComp struct {
	name string
}

func (ra testComp) Render(this *gr.This) gr.Component {
	return el.Div(gr.Text(ra.name))
}

var createHistory *js.Object

func forPath(path string) grouter.History {
	return grouter.History{createHistory.Invoke(path)}
}

func init() {

	reactRouter := js.Global.Get("ReactRouter")

	if reactRouter == js.Undefined {
		panic("Make sure that react-router is loaded, see https://github.com/reactjs/react-router")
	}

	// Memory history doesn't manipulate or read from the address bar.
	createHistory = reactRouter.Get("createMemoryHistory")

	if createHistory == js.Undefined {
		panic("ReactRouter.createHistory not found.")
	}

}
