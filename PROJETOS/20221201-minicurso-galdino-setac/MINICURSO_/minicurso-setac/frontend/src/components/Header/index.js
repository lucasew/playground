import React from "react";

import "./styles.css";

const Header = ({ history }) => {
  const handleLogout = () => {
    localStorage.removeItem("@access_token");
    history.push("/");
  };

  return (
    <div className="header">
      <p className="titlebar">
        Lanchonete dos <strong>ZAMIGOS</strong> | Visualização de Pedidos
      </p>
      <button type="button" onClick={handleLogout}>
        Sair
      </button>
    </div>
  );
};

export default Header;
