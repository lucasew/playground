#let footnote_state = state("footnote", ())
#let footnote_counter = counter("footnote")

#let footnote(text) = {
    locate(loc => {
        let counter = footnote_counter.at(loc)
        footnote_state.update(old => {
            old.push((
                index: counter.first(),
                text: text,
                page: loc.page()
            ))
            old
        })
        [ #super[ [#counter.first()] ] ]
    })
    footnote_counter.step()
}


#set page(
    paper: "a4",
    header: align(right, [
        #lorem(5)
    ]),
    footer: [
        /* Footnotes of this page */
        #locate(loc => {
            let state = footnote_state.at(loc)
            for x in state.filter(x => x.page == loc.page()) {
                [ #x.index. #x.text #linebreak() ]
            }
        })

        #counter(page).display()
    ],
    numbering: "1"
)

#let sum(x, y) = [A soma de #x e #y é #(x + y)]

= Teste
#sum(6, 9)

#lorem(50)
#footnote("Teste")
#lorem(100)
#footnote("Teste")
#lorem(50)

#pagebreak()

#lorem(100)
#footnote("Teste")
#lorem(100)

```python
import os
from sys import stdout
```

```nix
pkgs = import <nixpkgs> {};
```
