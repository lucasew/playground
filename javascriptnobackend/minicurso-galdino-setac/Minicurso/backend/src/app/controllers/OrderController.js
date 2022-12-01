const { Order } = require("../models/Order");
const { User } = require("../models/User");

class OrderController {
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

  async list(req, res) {
    const orders = await Order.find({ status: "IN PREPARATION" })
      .populate("products")
      .populate("attendant", "name");

    return res.json(orders);
  }

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
