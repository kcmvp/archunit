package archunit

type ArchRule interface {
	Names() []string
	Skip(names ...string) ArchRule
}
