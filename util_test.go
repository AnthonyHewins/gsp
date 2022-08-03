package gsp

import "github.com/bradleyjkemp/cupaloy/v2"

func snapshotTest() *cupaloy.Config {
	return cupaloy.NewDefaultConfig().WithOptions(cupaloy.UseStringerMethods(false))
}
