import { state } from "../../core/state.js";

// render function to display the login form and any error messages
export function render() {
  const app = document.querySelector("#app");

  app.innerHTML = `
    <div class="login-wrapper">
      <div class="card login-card">
        <h2>Login</h2>
        <form id="login-form" class="login-form-content">
          <label>
            Email
            <input id="email" type="email" placeholder="Email" required />
          </label>
          <label>
            Password
            <input id="password" type="password" placeholder="Password" required />
          </label>
          <button type="submit" ${state.auth.loading ? "disabled" : ""}>
            ${state.auth.loading ? "Logging in..." : "Login"}
          </button>
          ${state.auth.error ? `<div class="error-box">${state.auth.error}</div>` : ""}
        </form>
      </div>
    </div>
  `;
}
