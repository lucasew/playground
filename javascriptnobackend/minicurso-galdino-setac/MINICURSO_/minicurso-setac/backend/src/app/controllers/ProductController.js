const { Product } = require("../models/Product");

class ProductController {
  async store(req, res) {
    const { name, description, price, type } = req.body;

    const product = await Product.create({ name, description, price, type });

    return res.status(201).json(product);
  }
  async list(req, res) {
    const products = await Product.find();

    return res.json(products);
  }
}

module.exports = new ProductController();
