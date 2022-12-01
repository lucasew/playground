import api from "../utils/api";

import makeHeaders from "./sessionService";

const fetchOrdersFromAPI = async push => {
  // 1. Obtém o header de autenticação
  const headers = makeHeaders(push);

  try {
    // 3. Realizar a requisição
    const { data } = await api.get("/orders", { headers });

    // retornar os dados para quem me chamou
    return data;
  } catch (err) {
    alert(err);
  }
};

const closeOrder = async (orderId, push) => {
  // 1. Obtém o header de autenticação
  const headers = makeHeaders(push);

  try {
    await api.put(`/orders/${orderId}`, {}, { headers });
    return true;
  } catch (err) {
    alert(err);
    return false;
  }
};

export { fetchOrdersFromAPI, closeOrder };
