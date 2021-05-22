package main

import (
	"bytes"
	"fmt"

	"github.com/gomarkdown/markdown"
)

func main() {
    buf := bytes.NewBufferString("")
    fmt.Fprintln(buf, `
# Hello, world

<div id="test">
</div>
<style>
#test {
    width: 50px;
    height: 50px;
    background-color: red;
}
</style>
<script>
document.getElementById("test").innerText = "teste"
</script>
    `)
    fmt.Fprintln(buf, "```js")
    fmt.Fprintln(buf, `console.log("hello, world")`)
    fmt.Fprintln(buf, "```")
    data := buf.Bytes()
    println("Input")
    println(string(data))
    html := markdown.ToHTML(data, nil, nil)
    println("Output")
    println(string(html))
}
