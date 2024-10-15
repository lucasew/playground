import React, { useState, useEffect } from "react";

import { fetchOrdersFromAPI, closeOrder } from "../../services/orderService";

import "./styles.css";

import Header from "../../components/Header";
import Order from "../../components/Order";

const Dashboard = ({ history }) => {
  const [orders, setOrders] = useState([]);

  const populateState = async () => {
    const orders = await fetchOrdersFromAPI(history.push);
    setOrders(orders);
  };

  useEffect(() => {
    populateState();
  }, []);

  const handleCloseOrder = orderId => {
    if (closeOrder(orderId, history.push)) {
      let ordersUpdated = orders.filter(order => order._id !== orderId);

      setOrders(ordersUpdated);
    }
  };

  return (
    <div className="dashboard">
      <Header history={history} />

      <div className="content">
        {!orders.length && <h1>Não há pedidos</h1>}
        {orders.map(order => (
          <Order key={order._id} order={order} closeFn={handleCloseOrder} />
        ))}
      </div>
    </div>
  );
};

export default Dashboard;
