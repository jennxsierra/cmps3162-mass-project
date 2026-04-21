import { emitter } from "../../core/event-emitter.js";
import { state } from "../../core/state.js";
import { DataService } from "../../core/data-service.js";
import { render } from "./render.js";
import { setupHandlers } from "./handlers.js";

// auth:login -> call API
emitter.on("auth:login", ({ email, password }) => {
  state.auth.loading = true;
  state.auth.error = null;
  render();
  DataService.login({ email, password });
});

// auth:loginSuccess -> store token, redirect to appointments page
emitter.on("auth:loginSuccess", (token) => {
  state.auth.loading = false;
  state.auth.token = token;
  localStorage.setItem("auth_token", token);
  window.location.href = "/appointments";
});

// auth:loginError -> store error and rerender
emitter.on("auth:loginError", (msg) => {
  state.auth.loading = false;
  state.auth.error = msg;
  render();
});

// Initialize handlers and render initial state
setupHandlers();
render();
