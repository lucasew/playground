const jwt = require("jsonwebtoken");
const { promisify } = require("util");

const authConfig = require("../config/auth");
const { User } = require("../models/User");

/**
 * Express middleware to validate JWT authentication tokens.
 *
 * Extracts the token from the `Authorization` header, verifies its validity using the
 * configured secret, and securely attaches the decoded user ID (`req.userId`) to the
 * request object for consumption by subsequent downstream middleware or route handlers.
 *
 * @param {import('express').Request} req - The Express request object. It is mutated by attaching `req.userId`.
 * @param {import('express').Response} res - The Express response object. Used to send 401 Unauthorized if validation fails.
 * @param {import('express').NextFunction} next - The next middleware function in the stack.
 * @returns {Promise<void>}
 *
 * @throws Will not throw an unhandled exception, but terminates the request lifecycle
 * by sending a 401 response if the token is missing, malformed, or fails signature validation.
 */
module.exports = async (req, res, next) => {
  const authHeader = req.headers.authorization;

  if (!authHeader) {
    return res.status(401).json({ error: "Token not provided" });
  }

  const [, token] = authHeader.split(" ");

  try {
    const decoded = await promisify(jwt.verify)(token, authConfig.secret);

    req.userId = decoded.id;

    return next();
  } catch (err) {
    return res.status(401).json({ error: "Token invalid" });
  }
};
