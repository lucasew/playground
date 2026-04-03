package workspaced

import (
	"tool/cli"
	"tool/http"
)

command: robots: {
	fetch: http.Get & {
		url: "https://www.google.com/robots.txt"
	}

	print: cli.Print & {
		text: fetch.response.body
	}
}
