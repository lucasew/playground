import * as Server from "react-dom/server";
import Koa from "koa";
import Router from "@koa/router";
import RootComponent from "./App";
import esbuild from "esbuild";
import { join } from "path";

const rootFolder = __dirname;

const port = 8080;
const app = new Koa();

const r = new Router();
r.get("index.js", (ctx, next) => {
    esbuild.build({
        entryPoints: [join(rootFolder, 'client.tsx')],
    })
});
app.use(async (ctx) => {
  // This is vanilla SSR xD
  // BTW that print does not run on client
  ctx.body = Server.renderToString(<RootComponent />);
});
app.listen(port);
console.log(`listening on port ${port}`);
