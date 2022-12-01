const express = require("express");

const routes = express.Router();

/* Controllers */
const UserController = require("./app/controllers/UserController");
const SessionController = require("./app/controllers/SessionController");
const ProductController = require("./app/controllers/ProductController");
const OrderController = require("./app/controllers/OrderController");

/* Middlewares */
const PermissionMiddleware = require("./app/middlewares/PermissionMiddleware");
const AuthMiddleware = require("./app/middlewares/AuthMiddleware");

routes.get("/", (req, res) => {
  return res.send("Bem vindo ao aplicativo do minicurso!");
});

routes.post("/users", UserController.store);
routes.post("/session", SessionController.login);

routes.use(AuthMiddleware);
routes.post("/orders", OrderController.store);
routes.get("/products", ProductController.list);
routes.get("/orders", OrderController.list);

routes.use(PermissionMiddleware);
routes.post("/products", ProductController.store);
routes.put("/orders/:orderId", OrderController.update);

module.exports = routes;
