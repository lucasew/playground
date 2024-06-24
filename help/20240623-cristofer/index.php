<h1>Bucetola</h1>
<h2>Bucetola</h2>
<h3>Bucetola</h3>
<h4>Bucetola</h4>
<h5>Bucetola</h5>
<h6 id="sus">Bucetola</h6>

<ul>
    <li><a href="/gerar-curriculo">Gerar currículo</a></li>
    <li><a target="_blank" href="https://www.youtube.com/watch?v=dQw4w9WgXcQ">Documentação</a></li>
</ul>

<div id="divpreta" style="background-color: black; aspect-ratio: 1; max-width: 80vw; margin: auto">
    <h1 style="font-size: 420px; text-align: left">👀</h1>
</div>

<style>
#divpreta {
    display: flex;
    align-items: center;
    justify-content: center;
}
#divpreta h1 {
    opacity: 0;
    transition: all 2s;
}
#divpreta:hover h1 {
    opacity: 1;
}
</style>

<script>
document.getElementById("sus").addEventListener("click", () => {
    alert("There is a impostor among us")
})
</script>

<style>
.vermueilho {
    color: red;
}

h6:hover {
    display: none;
}

@media print {
    .vermueilho {
        color: blue;
    }
}
</style>

<p>
<span class="vermueilho">L</span>
<span color="green">G</span>
<span class="vermueilho">B</span>

<p>
<?php
parse_str($_SERVER["QUERY_STRING"], $qs);

var_dump($qs);
?>
</p>

<p>

<?php
echo "AAAAAAAA ";
echo 2+2;
?>
</p>

<p>
<?php
$coisa = "aaaaa";

$lugar = "coisa";
?>
</p>

<p>
<?php
echo $$lugar; // $coisa -> "aaaaa"

$$lugar = "bbbbb";
echo $coisa;

echo $_SERVER['QUERY_STRING'];
?>
</p>

<p>

T</p>

<script>
alert("você ganhou um iphone 16")
</script>

<script>
for (let i = 0; i < 10; i++) {
    console.log(i);
    // debugger;
}
</script>


<p>Nome: <span id="nome">Não definido</span></p>

<button onclick="document.getElementById('nome').innerText = prompt('Qual o nome?')">Alterar nome</button>
