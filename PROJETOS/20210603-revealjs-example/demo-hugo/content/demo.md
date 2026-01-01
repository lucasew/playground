---
title: Demonstração das funcionalidades
logo:
    src: https://blog-do-lucao.vercel.app/apple-touch-icon.png
    alt: Lucão
---

# Demonstração dos testes usando Reveal e Hugo

---

## Funciona a logo da Internet
- A logo é importada como asset então é redistribuida com o site gerado
- Tem que ser num formato mais padrão tipo PNG ou JPG senão o hugo reclama que não sabe lidar

----
%auto-animate%

## Funciona código

```java
public class Main {
    public static void main(String args[]) {
        System.out.println("aha");
    }
}
// Parece bem funcional pra mim
```

---
%auto-animate%

## Funciona código

```nix
# Svelte reclamava bastante desse trecho
{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
    buildInputs = with pkgs; [ hugo ];
}
```

---

## Funciona matemática

- $$x^2$$
- $$\frac{-b \pm \sqrt{b^2 -4ac}}{2a}$$

---

## Funciona HTML arbitrário
- Precisa ativar uma flag na configuração do hugo pra dar certo

<button onclick="alert('vai dizer que não')">Testar</button>

---
## Funciona embed

<iframe width="420" height="315" allowfullscreen src="https://www.youtube.com/embed/dQw4w9WgXcQ?autoplay=1&mute=0">
</iframe>

---
%auto-animate%
## Animações

<p>AAAA</p>

---
%auto-animate%
## Animações

<p color="red">AAAA</p>

---
<style>
img.block {
width: 100px; height: 100px;
margin: auto;
}
</style>

%auto-animate%

# Animações

<img class="block" style="background-color: red"></img>

---
%auto-animate%

# Animações

<img class="block" style="background-color: blue; margin-top: 100px"></img>

