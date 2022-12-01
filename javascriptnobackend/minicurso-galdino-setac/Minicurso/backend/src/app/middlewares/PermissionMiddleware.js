const { User } = require("../models/User");

module.exports = async (req, res, next) => {
  const user = await User.findById(req.userId);

  if (user.permission !== "ADMIN") {
    return res
      .status(401)
      .json({ error: "Only admins can perform that operation" });
  }

  return next();
};
