package regexparser

import (
	"fmt"
	"regexp"
	"strconv"
)

func ParseTmdbID(regex *regexp.Regexp, content string) (tmdbid int64, err error) {
	groups := regex.FindStringSubmatch(content)
	subnames := regex.SubexpNames()
	if len(groups) != len(subnames) {
		return 0, fmt.Errorf("tmdbid no match")
	}
	for i := 0; i < len(groups); i++ {
		if i == 0 {
			continue
		}
		subname := subnames[i]
		group := groups[i]
		switch subname {
		case "tmdbid":
			n, err := strconv.ParseInt(group, 10, 63)
			if err != nil {
				return 0, err
			}
			tmdbid = n
		}
	}
	if tmdbid <= 0 {
		return 0, fmt.Errorf("tmdbid invalid")
	}
	return tmdbid, nil
}
