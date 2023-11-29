package archunit

type Folders struct {
	selector []string
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
