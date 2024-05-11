// nolint
package archunit

import "github.com/samber/lo"

type Package lo.Tuple2[string, string]

func (pkg Package) ShouldNotRefer(path ...string) error {
	panic("implement me")
}

func (pkg Package) ShouldBeOnlyReferredBy(path ...string) error {
	panic("implement me")
}
