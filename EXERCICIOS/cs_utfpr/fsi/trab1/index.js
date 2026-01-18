const size_x = 10;
const size_y = 10;
const num_recompensas = 5;

function random_tile() {
  let tileTypes = [1, 10, 4, 20, 9999];
  return tileTypes[randint(tileTypes.length)];
}

function useCounter(until = 9999) {
  let i = 0;
  return () => {
    if (until <= i) {
      return null;
    }
    const ret = i;
    i++;
    return ret;
  };
}

function usePromiseDeferer(promiseFn) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      promiseFn().then(resolve).catch(reject);
    }, 1);
  });
}

function useTimeoutLoop(fn, delay = 200) {
  let stop = false;
  const cancel = () => {
    stop = true;
  };
  let timeouter = () => {
    if (stop) {
      return;
    }
    fn(cancel);
    setTimeout(timeouter, delay);
  };
  timeouter();
  return cancel;
}

function tile2label(tile) {
  return {
    1: "Sólido e plano",
    10: "Rochoso",
    4: "Arenoso",
    20: "SAIA AGORA DO MEU PÂNTANO",
    9999: "Não vai dar não",
  }[String(tile)];
}

function randint(to) {
  return Math.floor(to * Math.random());
}

function generate() {
  resetView();
  const [alvo_x, alvo_y] = [randint(size_x), randint(size_y)];
  let recompensas = [];
  for (let i = 0; i < num_recompensas; i++) {
    recompensas.push([randint(size_x), randint(size_y)]);
  }
  function isRecompensa(x, y) {
    for (let i = 0; i < num_recompensas; i++) {
      const [a, b] = recompensas[i];
      if (x === a && y === b) {
        return true;
      }
    }
    return false;
  }
  root = document.createElement("div");
  for (let i = 0; i < size_x; i++) {
    row = document.createElement("div");
    row.dataset.row = i;
    for (let j = 0; j < size_y; j++) {
      column = document.createElement("div");
      column.innerHTML = `<span>(${i}, ${j})</span>`;
      column.classList.add("local");
      const tile = random_tile();
      column.dataset.tile = tile;
      column.title = tile2label(tile);
      if (isRecompensa(i, j)) {
        column.dataset.recompensa = true;
      }
      if (i == alvo_x && j == alvo_y) {
        column.dataset.alvo = true;
      }
      column.dataset.row = i;
      column.dataset.column = j;
      row.appendChild(column);
    }
    root.appendChild(row);
  }
  document.getElementById("board").innerHTML = root.innerHTML;
  return root;
}

function getElements() {
  let ret = {};
  const board = document.getElementById("board");
  for (let i = 0; i < board.children.length; i++) {
    const this_row = board.children[i];
    for (let j = 0; j < this_row.children.length; j++) {
      const element = this_row.children[j];
      const { row, column, recompensa, alvo } = element.dataset;
      if (ret[row] === undefined) {
        ret[row] = {};
      }
      const custo = parseInt(element.dataset.tile);
      const isRecompensa = recompensa === "true";
      const isAlvo = alvo === "true";
      const isInacessivel = custo === 9999;
      ret[row][column] = {
        custo,
        isRecompensa,
        isAlvo,
        isInacessivel,
        element,
      };
    }
  }
  return new ElementsWrapper(ret, size_x, size_y);
}

generate();

class Queue {
  constructor() {
    this.items = [];
  }
  async add(item) {
    this.items.push(item);
  }
  remove() {
    return this.items.shift();
  }
  isEmpty() {
    return this.items.length == 0;
  }
  len() {
    return this.items.length;
  }
}

class ElementsWrapper {
  constructor(elements = getElements(), sizex = size_x, sizey = size_y) {
    this.elements = elements;
    this.size_x = sizex;
    this.size_y = sizey;
    this.continueOperation = true;
  }
  isIndiceValido(x, y) {
    return x < this.size_x && y < this.size_y && x >= 0 && y >= 0;
  }
  node(x, y) {
    if (!this.isIndiceValido(x, y)) {
      return null;
    }
    return this.elements[x][y];
  }
  heuristica(x, y) {
    const node = this.node(x, y);
    let ret = 0;
    if (node.isInacessivel) {
      return 0;
    }
    if (node.isAlvo) {
      ret += 20;
    }
    if (node.isRecompensa) {
      ret + 10;
    }
    return ret;
  }
  custo(x, y) {
    const node = this.node(x, y);
    return node.custo;
  }
  isInacessivel(x, y) {
    const node = this.node(x, y);
    return node.isInacessivel;
  }
  isRecompensaInacessivel(x, y) {
    const node = this.node(x, y);
    return this.isInacessivel(x, y) && node.isRecompensa;
  }
  isAlvoInacessivel(x, y) {
    const node = this.node(x, y);
    return this.isInacessivel(x, y) && node.isAlvo;
  }
  isRequeridoEInacessivel(x, y) {
    return this.isRecompensaInacessivel(x, y) && this.isAlvoInacessivel(x, y);
  }
  async expandNode(x, y) {
    const that = this;
    let candidates = [];
    [x - 1, x, x + 1].forEach((x) => {
      [y - 1, y, y + 1].forEach((y) => {
        candidates.push([x, y]);
      });
    });
    candidates = candidates.filter((v) => {
      const [nx, ny] = v;
      if (nx == x && ny == y) {
        return false;
      }
      return that.isIndiceValido(nx, ny) && !that.isInacessivel(nx, ny);
    });
    return candidates;
  }
  checkInitialState(start_x = 0, start_y = 0) {
    const that = this;
    if (that.isInacessivel(start_x, start_y)) {
      throw "Estado inicial inacessível. Problema sem solução :/.";
    }
  }
  async algoBFS(start_x = 0, start_y = 0) {
    this.continueOperation = true;
    this.checkInitialState(start_x, start_y);
    const that = this;
    let q = new Queue();
    await q.add([start_x, start_y, []]);
    async function handleItem(item) {
      const [x, y, path] = item;
      if (path.length > that.size_x * that.size_y) {
        throw "Estado final não encontrao";
      }
      if (that.isInacessivel(x, y)) {
        return null;
      }
      if (that.node(x, y).isAlvo) {
        return [x, y, path];
      }
      const neighbors = await that.expandNode(x, y);
      let newPath = [...path, [x, y]];
      for (let i = 0; i < neighbors.length; i++) {
        const [nx, ny] = neighbors[i];
        if (nx === undefined || ny === undefined) {
          continue;
        }
        if (
          newPath
            .map((item) => item.map(String).join(","))
            .includes(`${nx},${ny}`)
        ) {
          // checa se já passou por ali
          // console.log("found cycle, skipping")
          continue;
        }
        await q.add([nx, ny, newPath]);
      }
    }
    while (!q.isEmpty()) {
      // console.log("iteração")
      if (!that.continueOperation) {
        throw "Operação cancelada";
      }
      const item = await new Promise((resolve, reject) => {
        setTimeout(() => {
          // gambiarra pro processamento não travar tudo e poder cancelar
          handleItem(q.remove()).then(resolve).catch(reject);
        }, 1);
      });
      if (item) {
        return item;
      }
    }
    throw "Nenhum caminho encontrado";
  }
  async algoDFS(start_x = 0, start_y = 0) {
    this.checkInitialState(start_x, start_y);
    this.continueOperation = true;
    const that = this;
    async function recur(start_x = 0, start_y = 0, tail = []) {
      if (!that.continueOperation) {
        throw "Operação cancelada";
      }
      that.continueOperation = true;
      if (
        tail
          .map((v) => JSON.stringify(v))
          .includes(JSON.stringify([start_x, start_y]))
      ) {
        // console.log("found cycle, skipping...")
        return null;
      }
      const cur_node = that.node(start_x, start_y);
      if (cur_node.isInacessivel) {
        return null;
      }
      if (cur_node.isAlvo) {
        return [start_x, start_y, tail];
      }
      const neighbours = await that.expandNode(start_x, start_y);
      for (let i = 0; i < neighbours.length; i++) {
        const [x, y] = neighbours[i];
        const newTail = [...tail, [start_x, start_y]];
        const ret = await recur(x, y, newTail);
        if (ret == null) {
          continue;
        } else {
          return ret;
        }
      }
      throw "Nenhum caminho encontrado";
    }
    return await recur(start_x, start_y);
  }
  async algoAstar(start_x = 0, start_y = 0) {
    this.checkInitialState(start_x, start_y);
    const that = this;
    this.continueOperation = true;
    let melhor = null;
    let custoMelhor = 9998;
    async function recur(start_x = 0, start_y = 0, custo = 0, tail = []) {
      if (!that.continueOperation) {
        if (melhor != null) {
          return melhor;
        }
        throw "Operação cancelada";
      }
      that.continueOperation = true;
      that.checkInitialState(start_x, start_y);
      if (
        tail
          .map((v) => JSON.stringify(v))
          .includes(JSON.stringify([start_x, start_y]))
      ) {
        // console.log("found cycle, skipping...")
        return null;
      }
      const cur_node = that.node(start_x, start_y);
      if (cur_node.isInacessivel) {
        return null;
      }
      if (cur_node.isAlvo) {
        if (custo < custoMelhor) {
          console.log(`encontrado caminho com custo melhor: ${custo}`);
          custoMelhor = custo;
          melhor = [start_x, start_y, tail];
          document.getElementById("optimum").innerText =
            `${custoMelhor} (cancelar = mostrar)`;
        }
        return melhor;
      }
      const neighbours = await that.expandNode(start_x, start_y);
      let heuristicNeighbours = neighbours.sort((a, b) => {
        const [xa, ya] = a;
        const [xb, yb] = b;
        return (
          that.heuristica(xb, yb) -
          that.custo(xb, yb) -
          (that.heuristica(xa, ya) - that.custo(xa, ya))
        );
      });
      for (let i = 0; i < heuristicNeighbours.length; i++) {
        const [x, y] = neighbours[i];
        const nodeCusto = that.custo(x, y);
        const newTail = [...tail, [start_x, start_y]];
        const res = await usePromiseDeferer(() =>
          recur(x, y, custo + nodeCusto, newTail)
        );
        if (res != null) {
          return null;
        }
      }
    }
    await recur(start_x, start_y, this.custo(start_x, start_y));
    return melhor;
  }
  async algoGula(start_x = 0, start_y = 0) {
    const that = this;
    this.continueOperation = true;
    this.checkInitialState(start_x, start_y);
    async function recur(start_x = 0, start_y = 0, tail = []) {
      if (!that.continueOperation) {
        throw "Operação cancelada";
      }
      that.checkInitialState(start_x, start_y);
      if (
        tail
          .map((v) => JSON.stringify(v))
          .includes(JSON.stringify([start_x, start_y]))
      ) {
        // console.log("found cycle, skipping...")
        return null;
      }
      const cur_node = that.node(start_x, start_y);
      if (cur_node.isInacessivel) {
        return null;
      }
      if (cur_node.isAlvo) {
        return [start_x, start_y, tail];
      }
      const neighbours = await that.expandNode(start_x, start_y);
      let heuristicNeighbours = neighbours.sort((a, b) => {
        const [xa, ya] = a;
        const [xb, yb] = b;
        return that.heuristica(xb, yb) - that.heuristica(xa, ya);
      });
      for (let i = 0; i < heuristicNeighbours.length; i++) {
        const [x, y] = neighbours[i];
        const newTail = [...tail, [start_x, start_y]];
        const ret = await usePromiseDeferer(() => recur(x, y, newTail));
        if (ret == null) {
          continue;
        } else {
          return ret;
        }
      }
      throw "Nenhum caminho encontrado";
    }
    return await recur(start_x, start_y);
  }

  async highlightPath(data) {
    const that = this;
    if (data === null) {
      alert("Nenhum caminho viável foi encontrado!");
      return;
    }
    const [x, y, rest] = data;
    const path = [...rest, [x, y]];
    let custo = 0;
    path.forEach((v) => {
      const [x, y] = v;
      custo += that.custo(x, y);
    });
    document.getElementById("optimum").innerText = custo;
    const counter = useCounter(path.length);
    useTimeoutLoop((cancel) => {
      const i = counter();
      if (i == null) {
        cancel();
        return;
      }
      const [x, y] = path[i];
      const node = this.node(x, y);
      node.element.classList.add("path");
    });
  }
}

let elements = getElements();

function unmarkEveryone(className = "path") {
  while (true) {
    // não são todos os elementos que aparecem por rodada
    const elements = document.getElementsByClassName(className);
    if (elements.length == 0) {
      return;
    }
    for (let i = 0; i < elements.length; i++) {
      elements[i].classList.remove(className);
    }
  }
}

function resetView() {
  document.getElementById("optimum").innerText = "não conhecido";
  unmarkEveryone();
}

function elementsHandler(handler) {
  return async function () {
    resetView();
    elements = getElements();
    try {
      const res = await Promise.resolve(handler(elements));
      if (res == null) {
        throw "Nenhum caminho viável encontrado";
      }
      await elements.highlightPath(res);
      alert("Resultado encontrado com sucesso");
    } catch (e) {
      console.error(e);
      const msg = e.message || e;
      alert(msg);
    }
  };
}

const apply_busca_largura = elementsHandler((e) => e.algoBFS());
const apply_busca_profundidade = elementsHandler((e) => e.algoDFS());
const apply_busca_gulosa = elementsHandler((e) => e.algoGula());
const apply_busca_astar = elementsHandler((e) => e.algoAstar());

function cancel_operation() {
  if (elements) {
    console.log("cancelando operação");
    elements.continueOperation = false;
  }
}

setInterval(() => {
  console.log("teste de inanição");
}, 100);
