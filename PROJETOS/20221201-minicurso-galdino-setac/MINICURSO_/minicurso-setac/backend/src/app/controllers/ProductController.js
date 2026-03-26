const { Product } = require("../models/Product");

/**
 * Controller responsible for managing products.
 * Handles CRUD operations associated with menu items or inventory records.
 */
class ProductController {
  /**
   * Creates a new product entry in the database.
   * Stores product details like name, description, price, and category type.
   *
   * @param {import("express").Request} req - The Express request object containing `name`, `description`, `price`, and `type` in the body.
   * @param {import("express").Response} res - The Express response object.
   * @returns {Promise<import("express").Response>} Returns the newly created product document.
   */
  async store(req, res) {
    const { name, description, price, type } = req.body;

    const product = await Product.create({ name, description, price, type });

    return res.status(201).json(product);
  }
  /**
   * Retrieves all available products from the database.
   * Fetches the entire collection of products, typically used to display menu items or catalog data.
   *
   * @param {import("express").Request} req - The Express request object.
   * @param {import("express").Response} res - The Express response object.
   * @returns {Promise<import("express").Response>} Returns an array of all registered products.
   */
  async list(req, res) {
    const products = await Product.find();

    return res.json(products);
  }
}

module.exports = new ProductController();
