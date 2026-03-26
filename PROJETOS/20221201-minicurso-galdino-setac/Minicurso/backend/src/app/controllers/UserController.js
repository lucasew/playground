const { User } = require("../models/User");

/**
 * Controller responsible for managing user accounts.
 * Handles the registration flow, including automatic permission assignment
 * for the very first user created in the system.
 */
class UserController {
  /**
   * Registers a new user.
   * Checks for email uniqueness to avoid duplicate accounts.
   * If the database has no registered users yet, the new user will be granted
   * "ADMIN" privileges automatically. Otherwise, defaults to "ATTENDANT".
   *
   * @param {import("express").Request} req - The Express request object containing `email`, `name`, and `password` in the body.
   * @param {import("express").Response} res - The Express response object.
   * @returns {Promise<import("express").Response>} Returns the created user object along with a newly generated JWT token, or a 400 error if the email is already in use.
   */
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
