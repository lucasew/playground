const { User } = require("../models/User");

class UserController {
  async store(req, res) {
    // 1. Obtem o e-mail do corpo da requisição POST
    const { email } = req.body;

    // 2. Verifica se já existe algum usuário cadastrado com este e-email
    if (await User.findOne({ email })) {
      // 2.1 Se tiver, retorne uma mensagem de erro pro usuário
      return res.status(400).json({ error: "User already exists" });
    }

    let permission = "ATTENDANT";

    // 3. Verifica se é o primeiro usuário a se cadastrar na aplicação
    const users = await User.find();
    if (!users.length) {
      // 3.1 .find() retorna um array de usuários
      // se este array for vazio, então não há usuários na aplicação
      permission = "ADMIN";
    }

    // 3. Se não tiver, cadastre um novo usuário.
    // Corpo da requisição deve ter:
    // name
    // email
    // password
    const user = await User.create({ ...req.body, permission });

    // 4. Com o usuário criado, retorna esta informação a requisição
    return res.status(201).json({ user, token: User.generateToken(user) });
  }
}

module.exports = new UserController();
