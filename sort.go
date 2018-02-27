package main

type pagesByDate []*PageContext

func (s pagesByDate) Len() int {
	return len(s)
}

func (s pagesByDate) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s pagesByDate) Less(i, j int) bool {
	return s[i].Date.After(*s[j].Date)
}
