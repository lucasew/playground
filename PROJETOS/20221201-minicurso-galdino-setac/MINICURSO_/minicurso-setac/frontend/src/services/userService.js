import api from "../utils/api";

const logar = async (user, push) => {
  try {
    const { data } = await api.post("/session", user);

    localStorage.setItem("@access_token", data.token);

    push("/dashboard");
  } catch (err) {
    alert("Erro ao fazer login!");
  }
};

const cadastrar = async (user, push) => {
  try {
    const { data } = await api.post("/users", user);

    // Se o usuário for ADMIN
    if (data.user.permission === "ADMIN") {
      // Grave o token dele no localStorage
      localStorage.setItem("@access_token", data.token);

      // Mostre uma mensagem para o usuário
      alert(`Seja bem vindo Administrador ${data.user.name}`);

      // navegue-o para a tela de dashboard
      push("/dashboard");
    }

    // Se o usuário criado for atendente
    // Mostre uma mensagem informando alguns detalhes
    alert(
      `Conta do atendente ${data.user.name} foi criado. Faça login pelo aplicativo mobile.`
    );

    // navegue-o para a página principal :D
    push("/");
  } catch (err) {
    alert(err);
  }
};

export { logar, cadastrar };
