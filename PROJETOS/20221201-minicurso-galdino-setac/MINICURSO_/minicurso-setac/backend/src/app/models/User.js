const bcrypt = require("bcryptjs");
const jwt = require("jsonwebtoken");
const authConfig = require("../config/auth");
const mongoose = require("mongoose");

/**
 * Mongoose schema representing a registered user in the system.
 * It is used for both administrators and regular attendants.
 * Contains authentication data, timestamps, and an enum defining permission levels.
 *
 * @type {mongoose.Schema}
 */
const UserSchema = new mongoose.Schema({
  name: {
    type: String,
    required: true
  },
  email: {
    type: String,
    required: true,
    unique: true,
    lowercase: true
  },
  password: {
    type: String,
    required: true
  },
  createdAt: {
    type: Date,
    default: Date.now
  },
  permission: {
    type: String,
    enum: ["ADMIN", "ATTENDANT"],
    default: "ATTENDANT"
  }
});

/**
 * Pre-save hook to intercept document saving and apply a cryptographic hash to the user's password.
 * Only triggers if the password field is new or explicitly modified, preventing re-hashing of existing passwords.
 *
 * @param {Function} next - Callback to pass control to the next hook in the Mongoose lifecycle.
 */
UserSchema.pre("save", async function(next) {
  if (!this.isModified("password")) {
    return next();
  }

  this.password = await bcrypt.hash(this.password, 8);
});

UserSchema.methods = {
  /**
   * Compares an unencrypted password attempt against the user's hashed password.
   *
   * @param {string} password - The plain-text password to check.
   * @returns {Promise<boolean>} Resolves to true if the given password securely matches the hash.
   */
  compareHash(password) {
    return bcrypt.compare(password, this.password);
  }
};

UserSchema.statics = {
  /**
   * Generates a signed JSON Web Token using the system's auth secret.
   *
   * @param {Object} payload - Data payload to include inside the JWT.
   * @param {string} payload.id - The unique user database ID.
   * @returns {string} The resulting JWT token containing the payload and expiration config.
   */
  generateToken({ id }) {
    return jwt.sign({ id }, authConfig.secret, {
      expiresIn: authConfig.ttl
    });
  }
};

module.exports = {
  User: mongoose.model("User", UserSchema),
  UserSchema
};
