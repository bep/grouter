# go-react-router

[![GoDoc](https://godoc.org/github.com/bep/grouter?status.svg)](https://godoc.org/github.com/bep/grouter)


React-Router Bindings for GopherJS.

Will need https://github.com/bep/gr to run.

See it in action here: [http://bego.io/goreact/examples/router/](http://bego.io/goreact/examples/router/)

See also:

* https://github.com/reactjs/react-router
* https://cdnjs.com/libraries/react-router

## Example Setup

```go
var (
	component1   = gr.New(&clickCounter{title: "Counter 1", color: "#ff0066"})
	component2   = gr.New(&clickCounter{title: "Counter 2", color: "#339966"})
	component3   = gr.New(&clickCounter{title: "Counter 3", color: "#0099cc"})
	component3_2 = gr.New(&clickCounter{title: "Counter 3_2", color: "#ffcc66"})

	// WithRouter makes this.props.router happen.
	appComponent = gr.New(new(app), gr.Apply(grouter.WithRouter))

	router = grouter.New("/", appComponent,
		grouter.NewIndexRoute(grouter.Components{"main": component1}),
		grouter.NewRoute("c1", grouter.Components{"main": component1}),
		grouter.NewRoute("c2", grouter.Components{"main": component2}),
		grouter.NewRoute("c3", grouter.Components{"main": component3, "sub": component3_2}),
	)
)
```
