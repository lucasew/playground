import React, { useState } from "react";
import { cadastrar } from "../../services/userService";

import "./styles.css";

export default function Register({ history }) {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const validate = () => {
    if (!name || !name.trim) {
      alert("Nome inválido");
      return false;
    }
    if (!email || !email.trim) {
      alert("Email inválido");
      return false;
    }

    if (!password || !password.trim) {
      alert("Senha inválida");
      return false;
    }

    return true;
  };

  const handleRegister = async e => {
    e.preventDefault();

    if (validate()) {
      await cadastrar({ name, email, password }, history.push);
    }
  };

  return (
    <div className="container">
      <p className="title">
        Lanchonete dos <strong>ZAMIGOS</strong>
      </p>

      <h2>Cadastre-se</h2>

      <form className="sessionform" onSubmit={handleRegister}>
        <label className="form-label">Nome</label>
        <input
          type="text"
          value={name}
          onChange={e => setName(e.target.value)}
        />

        <label className="form-label">Email</label>
        <input
          type="email"
          value={email}
          onChange={e => setEmail(e.target.value)}
        />

        <label className="form-label">Senha</label>
        <input
          id="last"
          type="password"
          value={password}
          onChange={e => setPassword(e.target.value)}
        />

        <button type="submit" className="submit-btn">
          Cadastrar
        </button>
      </form>
    </div>
  );
}
