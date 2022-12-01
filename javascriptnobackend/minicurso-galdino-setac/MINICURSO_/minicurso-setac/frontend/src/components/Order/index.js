import React from "react";

import "./styles.css";

export default function Order({ order, closeFn }) {
  return (
    <div className="order">
      <strong>Mesa {order.table}</strong>
      <button type="button" onClick={() => closeFn(order._id)}>
        Fechar
      </button>
      <hr />

      <p>Pedidos:</p>
      <ul>
        {order.products.map(product => (
          <li key={`${product._id}-${Math.random()}`}>{product.name}</li>
        ))}
      </ul>
    </div>
  );
}
