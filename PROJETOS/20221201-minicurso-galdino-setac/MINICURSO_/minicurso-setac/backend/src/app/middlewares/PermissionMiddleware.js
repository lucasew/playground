const { User } = require("../models/User");

/**
 * Express middleware to enforce administrator access control.
 *
 * Consults the database to retrieve the user's role using `req.userId`
 * (which must be previously attached to the request object by `AuthMiddleware`).
 * Prevents unauthorized access by returning a 401 response if the user
 * lacks the `ADMIN` permission role.
 *
 * @param {import('express').Request} req - The Express request object, strictly requiring `req.userId`.
 * @param {import('express').Response} res - The Express response object, used to return 401 Unauthorized for non-admins.
 * @param {import('express').NextFunction} next - The next middleware function in the stack.
 * @returns {Promise<void>}
 *
 * @throws Will bubble unhandled exceptions from the DB layer (e.g. `User.findById`),
 * but explicitly terminates the request lifecycle with 401 if the user is not an administrator.
 */
module.exports = async (req, res, next) => {
  const user = await User.findById(req.userId);

  if (user.permission !== "ADMIN") {
    return res
      .status(401)
      .json({ error: "Only admins can perform that operation" });
  }

  return next();
};
