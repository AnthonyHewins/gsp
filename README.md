# GSP

`gsp` leverages generics to create struct promises:

```go
type PromiseAll struct {
	DBField int
	APIField string
}

func (p PromiseAll) Jobs() []pipeline.Job[string] {
	return []pipeline.Job{
		&pipeline.Synchronous[string]{
			
		},
	}
}

```

### Install

``` shell
go get -u github.com/AnthonyHewins/gsp
```
