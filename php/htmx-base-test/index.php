<html>
    <head>
        <title>Teste</title>

    </head>

    <body>
        <button hx-post="a.php" hx-swap="innerHTML" hx-target="#space">A</button>
        <button hx-post="b.php" hx-swap="innerHTML" hx-target="#space">B</button>
        <button hx-post="c.php" hx-swap="innerHTML" hx-target="#space">C</button>
        <div id="space"></div>
        <script src="https://unpkg.com/htmx.org@1.9.6"></script>
    </body>
</html>
