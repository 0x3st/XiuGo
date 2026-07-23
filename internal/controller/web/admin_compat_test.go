package web

import "testing"

func TestParseXiunoAdminRoute(t *testing.T) {
	tests := []struct {
		raw       string
		route     string
		action    string
		arguments []string
	}{
		{raw: "", route: "index"},
		{raw: "setting-base.htm", route: "setting", action: "base"},
		{raw: "forum-update-1.htm", route: "forum", action: "update", arguments: []string{"1"}},
		{
			raw: "user-list-username-a_2db_5fc-2.htm", route: "user", action: "list",
			arguments: []string{"username", "a-b_c", "2"},
		},
	}
	for _, test := range tests {
		route, action, arguments := parseXiunoAdminRoute(test.raw)
		if route != test.route || action != test.action || !equalStrings(arguments, test.arguments) {
			t.Fatalf("parse %q = %q, %q, %#v", test.raw, route, action, arguments)
		}
	}
}

func TestRunlevelAllowsMatchesOriginalRules(t *testing.T) {
	tests := []struct {
		level    int
		loggedIn bool
		method   string
		want     bool
	}{
		{level: 0, loggedIn: true, method: "GET", want: false},
		{level: 1, loggedIn: true, method: "GET", want: false},
		{level: 2, loggedIn: true, method: "GET", want: true},
		{level: 2, loggedIn: true, method: "POST", want: false},
		{level: 2, loggedIn: false, method: "GET", want: false},
		{level: 3, loggedIn: true, method: "POST", want: true},
		{level: 3, loggedIn: false, method: "GET", want: false},
		{level: 4, loggedIn: false, method: "GET", want: true},
		{level: 4, loggedIn: true, method: "POST", want: false},
		{level: 5, loggedIn: false, method: "POST", want: true},
	}
	for _, test := range tests {
		if got := runlevelAllows(test.level, test.loggedIn, test.method); got != test.want {
			t.Fatalf("runlevelAllows(%d, %v, %s) = %v, want %v", test.level, test.loggedIn, test.method, got, test.want)
		}
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
