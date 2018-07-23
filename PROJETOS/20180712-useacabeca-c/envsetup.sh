OUTFILE=/tmp/programa
function rodarc () {
    gcc $1 -o $OUTFILE -Wall && $OUTFILE
}

function buildc () {
    gcc $* -o $OUTFILE -Wall && echo "Seu programa está situado em $OUTFILE."
}

funcion buildgo () {
    go build -o $OUTFILE $* && echo "Seu programa está situado em $OUTFILE"
}
