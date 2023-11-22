package archunit

var _ NameRule = (*Folders)(nil)

type Folders struct {
	selector []string
}

func (folders *Folders) NameShouldContain(rule PartChecker, part string) error {
	panic("")
}

func (folders *Folders) NameCaseShouldBe(caseChecker CaseChecker) error {
	//TODO implement me
	panic("implement me")
}

func FolderWith(name ...string) *Folders {

	return nil
}

func AllFolders() *Folders {
	return nil
}

func (folders *Folders) Skip() *Folders {
	return nil
}
