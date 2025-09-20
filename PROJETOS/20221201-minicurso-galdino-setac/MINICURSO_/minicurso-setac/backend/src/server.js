require("dotenv").config({ path: ".env" });

const express = require("express");
const Morgan = require("morgan");
const BodyParser = require("body-parser");
const mongoose = require("mongoose");
const cors = require("cors");

// Express REST routes
const routes = require("./routes");

// Configurações do banco mongodb atlas
const databaseConfig = require("./app/config/database");

class Server {
  constructor() {
    this.webApp = express();

    this.database();
    this.middlewares();
    this.routes();
  }

  database() {
    mongoose.connect(databaseConfig.uri, {
      useCreateIndex: true,
      useNewUrlParser: true,
      useUnifiedTopology: true
    });
  }

  middlewares() {
    this.webApp.use(cors());

    // REST Loggin with Morgan
    this.webApp.use(
      Morgan(function(tokens, req, res) {
        return [
          tokens.method(req, res),
          tokens.url(req, res),
          tokens.status(req, res),
          tokens.res(req, res, "content-length"),
          "-",
          tokens["response-time"](req, res),
          "ms"
        ].join(" ");
      })
    );

    // Parse JSON with body-parser
    this.webApp.use(BodyParser.json());
  }

  routes() {
    this.webApp.use(routes);
  }
}

module.exports = new Server().webApp;
