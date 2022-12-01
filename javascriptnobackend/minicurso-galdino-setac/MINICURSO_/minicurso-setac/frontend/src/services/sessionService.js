const makeHeaders = push => {
  // 1. Obter o token de autênticação do localStorage
  const token = localStorage.getItem("@access_token");

  // Se não tiver este token no localStorage...
  if (!token) {
    // Forçar a autenticação do administrador
    push("/");
    return;
  }

  // 2. Montar o cabeçalho da requisição - autenticação
  const headers = {
    Authorization: `Bearer ${token}`
  };

  return headers;
};

export default makeHeaders;
