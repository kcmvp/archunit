// nolint
package archunit

type Type struct {
	name      string
	functions []Function
}

type Types []Type

func (types Type) Exclude(names ...string) Types {
	panic("to be implemented")
}

func (types Types) EmbeddedWith(embeds ...string) Types {
	panic("to be implemented")
}

func (types Types) Implement(inters ...string) Types {
	panic("to be implemented")
}

func (types Types) InPackage(paths ...string) Types {
	panic("to be implemented")
}

func (types Types) Functions() []Function {
	panic("to be implemented")
}

func (types Types) ShouldBePrivate() error {
	panic("to be implemented")
}

func (functions Types) ShouldBePublic() error {
	panic("to be implemented")
}

func (types Types) ShouldBeInPackages(pkgs ...string) error {
	panic("to be implemented")
}

func (types Types) NameShould(pattern NamePattern) error {
	panic("to be implemented")
}
