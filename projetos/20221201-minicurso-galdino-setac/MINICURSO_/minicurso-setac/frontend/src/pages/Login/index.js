import React, { useState, useEffect } from "react";
import { logar } from "../../services/userService";

import "./styles.css";

export default function Login({ history }) {
  const wasUserLogged = () => {
    if (localStorage.getItem("@access_token")) {
      history.push("/dashboard");
    }
  };

  useEffect(() => {
    wasUserLogged();
  });

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const validate = () => {
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

  const handleLogin = async e => {
    e.preventDefault();

    if (validate()) {
      await logar({ email: email, password: password }, history.push);
    }
  };

  return (
    <div className="container">
      <p className="title">
        Lanchonete dos <strong>ZAMIGOS</strong>
      </p>

      <h2>Login</h2>

      <form className="sessionform" onSubmit={handleLogin}>
        <label className="form-label">Usuário</label>
        <input
          id="first"
          type="email"
          value={email}
          onChange={e => setEmail(e.target.value)}
        />

        <label className="form-label">Senha</label>
        <input
          type="password"
          value={password}
          onChange={e => setPassword(e.target.value)}
        />

        <button type="submit" className="submit-btn">
          Entrar
        </button>
      </form>
    </div>
  );
}
