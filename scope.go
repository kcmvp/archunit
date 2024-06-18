package archunit

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/samber/lo"
)

func ScopePattern(paths ...string) ([]*regexp.Regexp, error) {
	pps := lo.FlatMap(paths, func(item string, _ int) []string {
		path := strings.TrimPrefix(strings.TrimSuffix(item, "/"), "/")
		return lo.Union([]string{path, strings.TrimSuffix(path, "/...")})
	})
	pattern := `^(?:[a-zA-Z]+(?:\.[a-zA-Z]+)*|\.\.\.)$`
	re := regexp.MustCompile(pattern)
	for _, path := range pps {
		for _, seg := range strings.Split(path, "/") {
			if len(seg) > 0 && !re.MatchString(seg) {
				return nil, fmt.Errorf("invalid package paths: %s", path)
			}
		}
	}
	return lo.Map(pps, func(path string, _ int) *regexp.Regexp {
		return regexp.MustCompile(fmt.Sprintf("%s$", strings.ReplaceAll(path, "...", ".*")))
	}), nil
}
