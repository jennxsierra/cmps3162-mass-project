import { state } from "../../core/state.js";

// render function to display the login form and any error messages
export function render() {
  const app = document.querySelector("#app");

  app.innerHTML = `
    <h2>Login</h2>
    <form id="login-form">
      <input id="email" type="email" placeholder="Email" required />
      <input id="password" type="password" placeholder="Password" required />
      <button type="submit" ${state.auth.loading ? "disabled" : ""}>
        ${state.auth.loading ? "Logging in..." : "Login"}
      </button>
      ${state.auth.error ? `<div class="error-box">${state.auth.error}</div>` : ""}
    </form>
  `;
}
