# GSP

`gsp` leverages generics to create struct promises. In the simplest form:

```go
type PromiseAll struct {
	FullName string
	BirthdayToday bool
}

var arguments = map[string]any{
	"username": "123",
	"password": "123",
	"birthday": "1/1/1970"
}

func main() {
	promiseAll, err := gsp.New[PromiseAll](arguments)
	if err != nil {
		panic(err)
	}
	
	fmt.Println(promiseAll.FullName) // prints FullName from first pipeline worker below
	fmt.Println(promiseAll.BirthdayToday) // prints BirthdayToday from second pipeline worker below
}

type pipelineWorker = gsp.Synchronous[map[string]any]

func (p PromiseAll) Jobs() []gsp.Job[string] {
	return []gsp.Job{
		&pipelineWorker{
			Fn: func(ctx context.Context, pipe chan<- gsp.Atom, errChan chan<- error, args map[string]any) (gsp.Atom, error) {
				resp, err := someDBCall(args["username"], args["password"])
				if err != nil {
					return nil, err
				} else if resp.FullName == "" {
					return nil, fmt.Errorf("user not found")
				}
				
				return gsp.Atom{Index: 0, Data: resp.FullName}, nil
			},
		},
		&pipelineWorker{
			Fn: func(ctx context.Context, pipe chan<- gsp.Atom, errChan chan<- error, args map[string]any) (gsp.Atom, error) {
				resp, err := someAPICall(args["birthday"])
				if err != nil {
					return nil, err
				}
				
				return gsp.Atom{Index: 1, Data: resp.BirthdayToday}, nil
			},
		},
	}
}
```

### Install

``` shell
go get -u github.com/AnthonyHewins/gsp
```
