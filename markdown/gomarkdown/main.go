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

$$
\left[ \begin{array}{a} a^l_1 \\ ⋮ \\ a^l_{d_l} \end{array}\right]
= \sigma(
 \left[ \begin{matrix}
  w^l_{1,1} & ⋯  & w^l_{1,d_{l-1}} \\
  ⋮ & ⋱  & ⋮  \\
  w^l_{d_l,1} & ⋯  & w^l_{d_l,d_{l-1}} \\
 \end{matrix}\right]  ·
 \left[ \begin{array}{x} a^{l-1}_1 \\ ⋮ \\ ⋮ \\ a^{l-1}_{d_{l-1}} \end{array}\right] +
 \left[ \begin{array}{b} b^l_1 \\ ⋮ \\ b^l_{d_l} \end{array}\right])
 $$

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
