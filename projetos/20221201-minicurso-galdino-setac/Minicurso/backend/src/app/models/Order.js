const mongoose = require("mongoose");

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
