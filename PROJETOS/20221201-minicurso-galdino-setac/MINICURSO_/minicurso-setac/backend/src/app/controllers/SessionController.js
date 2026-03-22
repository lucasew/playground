const { User } = require("../models/User");

/**
 * Controller responsible for authenticating users.
 * Handles login flows and generates JWT tokens for authenticated sessions.
 */
class SessionController {
  /**
   * Authenticates a user based on their email and password.
   * Validates the provided credentials and generates a signed JWT token
   * if the login is successful, which can be used to authenticate subsequent requests.
   *
   * @param {import("express").Request} req - The Express request object containing `email` and `password` in the body.
   * @param {import("express").Response} res - The Express response object.
   * @returns {Promise<import("express").Response>} Returns the authenticated user object and a JWT token, or a 400 error for invalid credentials.
   */
  async login(req, res) {
    const { email, password } = req.body;

    const user = await User.findOne({ email });

    if (!user) {
      return res.status(400).json({ error: "User not found" });
    }

    if (!(await user.compareHash(password))) {
      return res.status(400).json({ error: "Invalid password" });
    }

    return res.json({ user, token: User.generateToken(user) });
  }
}

module.exports = new SessionController();
