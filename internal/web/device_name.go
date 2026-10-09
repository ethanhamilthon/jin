package web

import "strings"

// deviceName turns a User-Agent into a short name such as "iPhone, Safari".
func deviceName(agent string) string {
	return firstMatch(agent, [][2]string{
		{"iPhone", "iPhone"}, {"iPad", "iPad"}, {"Android", "Android"},
		{"Macintosh", "Mac"}, {"Windows", "Windows"}, {"Linux", "Linux"},
	}, "Unknown device") + ", " + firstMatch(agent, [][2]string{
		{"Edg/", "Edge"}, {"OPR/", "Opera"}, {"Firefox", "Firefox"}, {"CriOS", "Chrome"},
		{"Chrome", "Chrome"}, {"Safari", "Safari"},
	}, "browser")
}

func firstMatch(agent string, names [][2]string, fallback string) string {
	for _, n := range names {
		if strings.Contains(agent, n[0]) {
			return n[1]
		}
	}
	return fallback
}
