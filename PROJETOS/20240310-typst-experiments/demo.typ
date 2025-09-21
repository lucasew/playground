#set par(justify: true, leading: 0.52em)
#set text(font: "Linux Libertine", size: 12pt)

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
    paper: "us-letter",
    header: align(right, [
        #lorem(5)
    ]),
    footer: [
        #locate(loc => {
            let state = footnote_state.at(loc)
            for x in state.filter(x => x.page == loc.page()) {
                [ #x.index. #x.text #linebreak() ]
            }
        })

        Opa #counter(page).display()
    ],
    numbering: "1"
)
#set heading(numbering: "1.a")


#show "ipsum": name => box[
    #box[
        #image("//home/lucasew/.icons/tixati.png", height: 0.7em)
    ]
    #name
]


#show heading: it => [
  #set align(center)
  #set text(16pt, weight: "regular")
  #block(smallcaps(it.body))
]

#let sum(x, y) = [A soma de #x e #y é #(x + y)]

= Teste
#sum(6, 9)

#footnote("Teste")
#footnote("Teste")
#footnote("Teste")

Tixati é um programa. Teste tixati.

#lorem(20)

#columns(2, [
== Teste

#lorem(20)

#figure(
    image("//home/lucasew/.icons/tixati.png", width: 50%),
) <favicon>
])

#pagebreak()

#footnote("Teste")

== Eoq

#counter(page).display()

$x^2$

#lorem(20)

#lorem(20)

```python
import os
from sys import stdout
```

```nix
pkgs = import <nixpkgs> {};
```

```json
{
    "name": "Lucas",
}
```
