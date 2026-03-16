const mongoose = require("mongoose");

/**
 * Mongoose schema representing a product available for order in the system.
 * It tracks core product details like name, description, price, and categorizes them into snacks or drinks.
 *
 * @type {mongoose.Schema}
 */
const ProductSchema = new mongoose.Schema({
  name: {
    type: String,
    required: true
  },
  description: {
    type: String,
    required: true
  },
  price: {
    type: Number,
    required: true
  },
  type: {
    type: String,
    enum: ["snack", "drink"],
    required: true
  },
  createdAt: {
    type: Date,
    default: Date.now
  }
});

module.exports = {
  Product: mongoose.model("Product", ProductSchema),
  ProductSchema
};
