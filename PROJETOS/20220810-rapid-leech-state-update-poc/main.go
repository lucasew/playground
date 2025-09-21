package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
    http.HandleFunc("/", handler)
    err := http.ListenAndServe(":42069", nil)
    if err != nil {
        panic(err)
    }
}

const demoHTML = `
<html>
<head>
    <title>Teste</title>
</head>
<body>
<script>
function changeCounter(str) {
    document.getElementById("counter").textContent = str
}
</script>
<h1 id="counter"></h1>
</body>
</html>
`

func handler(w http.ResponseWriter, r *http.Request) {
    defer println("Conexão encerrada")
    w.WriteHeader(200)
    // w.Header().Add("Connection", "keep-alive")
    // w.Header().Add("Keep-Alive", "timeout=5,max=1")
    w.Write([]byte(demoHTML))
    counter := 0
    ticker := time.Tick(time.Second)
    for {
        if f, ok := w.(http.Flusher); ok {
            f.Flush()
        }
        select {
        case <-ticker:
            fmt.Fprintf(w, "<script>changeCounter('%s')</script>", fmt.Sprintf("Contador: %d", counter))
            counter++
        case <-r.Context().Done():
            return
        }
        // if counter == 10 {
        //     break
        // }
    }
}
