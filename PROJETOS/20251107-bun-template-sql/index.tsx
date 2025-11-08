
function h(
  tag: string | Function,
  props: Record<string, any> | null,
  ...children: any[]
): string {
  if (typeof tag === 'function') {
    return tag({ ...props, children });
  }

  const attrs = props
    ? Object.entries(props)
        .map(([key, value]) => ` ${key}="${value}"`)
        .join('')
    : '';

  const childrenStr = children.flat(Infinity).join('');

  return `<${tag}${attrs}>${childrenStr}</${tag}>`;
}


import { SQL } from "bun"
const sql = new SQL("sqlite://:memory:")

const stuff = await sql`SELECT 1 as poggers`

console.log(stuff)

//const h =( x, y, z )=> [x,y,z]
const stuff2 = (
  <div id="teste">
    <h1>Teste</h1>
  </div>
)
console.log(stuff2)
