const { Order } = require("../models/Order");
const { User } = require("../models/User");

/**
 * Controller handling operations for customer orders.
 * Manages the lifecycle of an order from creation by an attendant
 * to listing active orders, and finally updating them to completed status.
 */
class OrderController {
  /**
   * Creates a new order assigned to a specific table.
   * Associates the order with the currently authenticated attendant making the request,
   * registers the requested products, and returns the populated order details.
   *
   * @param {import("express").Request} req - The Express request object. `req.userId` must be populated by the AuthMiddleware, and the body must contain `products` and `table`.
   * @param {import("express").Response} res - The Express response object.
   * @returns {Promise<import("express").Response>} Returns the newly generated order with populated product details.
   */
  async store(req, res) {
    // 1. Obtém os dados necessários para criar uma comanda
    const { products, table } = req.body;

    // 2. Busca o atendente que fez a comanda
    const attendant = req.userId;

    // 3. Cria a comanda
    const order = await Order.create({ attendant, products, table });

    //4. Populate
    await order.populate("products").execPopulate();

    // 4. Retorna a nova comanda
    return res.status(201).json(order);
  }

  /**
   * Fetches the current list of pending active orders.
   * Filters orders by "IN PREPARATION" status to present a view for attendants or kitchen staff.
   * Populates referenced products and the name of the attendant who submitted the order.
   *
   * @param {import("express").Request} req - The Express request object.
   * @param {import("express").Response} res - The Express response object.
   * @returns {Promise<import("express").Response>} Returns an array of orders in preparation.
   */
  async list(req, res) {
    const orders = await Order.find({ status: "IN PREPARATION" })
      .populate("products")
      .populate("attendant", "name");

    return res.json(orders);
  }

  /**
   * Completes an existing order.
   * Finds an order by its ID from request params and updates its state to "DONE".
   * Use this endpoint when an order is finalized and delivered to the table.
   *
   * @param {import("express").Request} req - The Express request object containing the `orderId` parameter.
   * @param {import("express").Response} res - The Express response object.
   * @returns {Promise<import("express").Response>} Returns the updated order object or a 404 error if not found.
   */
  async update(req, res) {
    // 1. Obtém o ID da ordem do corpo da requisição
    const { orderId } = req.params;

    console.log(orderId);

    // 2. Busca essa ordem
    const order = await Order.findById(orderId);

    console.log(order);

    // 3. Se não existir, retornar uma mensagem de erro
    if (!order) {
      return res.status(404).json({ error: "Order not found! " });
    }

    // 4. Alterar o status dessa ordem
    order.status = "DONE";

    // 5. Salvar esta ordem
    await order.save();

    return res.json(order);
  }
}

module.exports = new OrderController();
