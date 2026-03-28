const mongoose = require("mongoose");

/**
 * Mongoose schema representing an active or completed order in the system.
 * Orders link multiple products to the attendant handling them, and track the status of the order lifecycle.
 *
 * @type {mongoose.Schema}
 */
const OrderSchema = new mongoose.Schema({
  attendant: {
    type: mongoose.SchemaTypes.ObjectId,
    ref: "User"
  },
  products: [
    {
      type: mongoose.SchemaTypes.ObjectId,
      ref: "Product"
    }
  ],
  status: {
    type: String,
    enum: ["IN PREPARATION", "DONE"],
    default: "IN PREPARATION"
  },
  table: {
    type: String,
    required: true
  },
  createdAt: {
    type: Date,
    default: Date.now
  }
});

module.exports = {
  Order: mongoose.model("Order", OrderSchema),
  OrderSchema
};
