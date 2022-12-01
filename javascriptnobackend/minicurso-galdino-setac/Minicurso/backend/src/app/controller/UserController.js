const {User} = require('../models/User');

class UserController {
    async store(req, res) {
        // 1. Pegar o email da requisição POST
        const {email} = req.body;
        // 2. Verificar se o e-mail existe no BD
        if (await User.findOne({ email })) {
            // 2.1. Se tiver usuário, retornar erro
            return res
                .status(400)
                .json({error: "Usuário já existe"})
        }

        let permission = "ATTENDANT";
        // 3. Verificar se primeiro usuário
        const users = await User.find();
        if (!users.length) {
            permission = "ADMIN";
        }
        const user = await User.create({ ... req.body, permission: permission })
        // 4. Com o usuário criado, retorna para a requisição
        return res.status(201).json({
            user, 
            token: User.generateToken(user)
        })
    }
}

module.exports = new UserController();