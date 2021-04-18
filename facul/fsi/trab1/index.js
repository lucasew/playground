function random_tile() {
    let tileTypes = [1, 10, 4, 20, 9999]
    return tileTypes[randint(tileTypes.length)]
}

function tile2label(tile) {
    return {
        "1": "Sólido e plano",
        "10": "Rochoso",
        "4": "Arenoso",
        "20": "SAIA AGORA DO MEU PÂNTANO",
        "9999": "Não vai dar não"
    }[String(tile)]
}

function randint(to) {
    return Math.floor(to*Math.random())
}

const size_x = 10
const size_y = 10
const num_recompensas = 5
function generate() {
    const [alvo_x, alvo_y] = [randint(size_x), randint(size_y)]
    let recompensas = []
    for (let i = 0; i < num_recompensas; i++) {
        recompensas.push([randint(size_x), randint(size_y)])
    }
    function isRecompensa(x, y) {
        for (let i = 0; i < num_recompensas; i++) {
            const [a, b] = recompensas[i]
            if (x === a && y === b) {
                return true
            }
        }
        return false
    }
    root = document.createElement('div')
    for (let i = 0; i < size_x; i++) {
        row = document.createElement('div')
        row.dataset.row = i
        for (let j = 0; j < size_y; j++) {
            column = document.createElement('div')
            column.innerHTML = `<span>(${i}, ${j})\nDestino</span>`
            column.classList.add("local")
            const tile = random_tile()
            column.dataset.tile = tile
            column.title = tile2label(tile)
            if (isRecompensa(i, j)) {
                column.dataset.recompensa = true
            }
            if (i == alvo_x && j == alvo_y) {
                column.dataset.alvo = true
            }
            column.dataset.row = i
            column.dataset.column = j
            row.appendChild(column)
        }
        root.appendChild(row)
    }
    document.getElementById("board").innerHTML = root.innerHTML
    return root
}

function getElements() {
    let ret = {}
    const board = document.getElementById("board")
    for(let i = 0; i < board.children.length; i++) {
        const this_row = board.children[i]
        for (let j = 0; j < this_row.children.length; j++) {
            const element = this_row.children[j]
            const {row, column, recompensa} = element.dataset
            if (ret[row] === undefined) {
                ret[row] = {}
            }
            ret[row][column] = {
                custo: parseInt(element.dataset.tile),
                isRecompensa: recompensa === "true"
            }
        }
    }
    return ret
}

generate()
