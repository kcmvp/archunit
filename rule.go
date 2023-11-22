package archunit

type NameRule interface {
	NameShouldContain(partChecker PartChecker, part string) error
	NameCaseShouldBe(caseChecker CaseChecker) error
}
type PartChecker func(a, b string) error

type CaseChecker func(a string) error

func HasPrefix(name, prefix string) error {
	return nil
}
func HasSuffix(name, suffix string) error {
	return nil
}

func Contains(name, part string) error {
	return nil
}

func UpperCase(a string) error {
	return nil
}

func LowerCase(a string) error {
	return nil
}
