import { emitter } from "../../core/event-emitter.js";
import { state } from "../../core/state.js";
import { DataService } from "../../core/data-service.js";

import { render } from "./render.js";
import { setupHandlers } from "./handlers.js";

// Require token (your route needs appointments:read permission)
if (!state.auth.token) {
  window.location.href = "/login";
}

// appointments:fetch -> call API
emitter.on("appointments:fetch", () => {
  state.appointments.loading = true;
  state.appointments.error = null;
  render();

  DataService.fetchAppointments();
});

// appointments:fetched -> update state and rerender
emitter.on("appointments:fetched", (payload) => {
  state.appointments.loading = false;

  state.appointments.data = payload.appointments || [];
  state.appointments.metadata = payload["@metadata"] || {};

  render();
});

// appointments:pageChanged -> update filters.page then fetch again
emitter.on("appointments:pageChanged", (page) => {
  state.appointments.filters.page = page;
  emitter.emit("appointments:fetch");
});

// appointments:error -> store error and rerender
emitter.on("appointments:error", (msg) => {
  state.appointments.loading = false;
  state.appointments.error = msg;
  render();
});

// Initialize handlers + first fetch
setupHandlers();
render();
emitter.emit("appointments:fetch");
