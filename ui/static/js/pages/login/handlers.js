import { emitter } from "../../core/event-emitter.js";

// setupHandlers registers event listeners for user interactions on the login page
export function setupHandlers() {
  const app = document.querySelector("#app");

  app.addEventListener("submit", (e) => {
    if (e.target.id !== "login-form") return;
    e.preventDefault();

    emitter.emit("auth:login", {
      email: document.querySelector("#email").value.trim(),
      password: document.querySelector("#password").value,
    });
  });
}
