import { emitter } from "../../core/event-emitter.js";
import { state } from "../../core/state.js";

export function setupHandlers() {
  const app = document.querySelector("#app");

  // Pagination buttons (event delegation)
  app.addEventListener("click", (e) => {
    const btn = e.target.closest("[data-page]");
    if (!btn) return;

    const page = Number(btn.dataset.page);
    if (!Number.isFinite(page) || page < 1) return;

    emitter.emit("appointments:pageChanged", page);
  });

  // Apply filters
  app.addEventListener("click", (e) => {
    if (e.target.id !== "apply-filters") return;

    const f = state.appointments.filters;

    f.include_cancelled = document.querySelector(
      "#filter-include-cancelled",
    ).checked;

    f.page_size = Number(document.querySelector("#filter-page-size").value);
    f.sort = document.querySelector("#filter-sort").value;

    f.provider_id = "";
    f.patient_id = "";
    f.appt_type_id = "";
    f.start_from = "";
    f.start_to = "";

    // reset to page 1 whenever filters change
    f.page = 1;

    emitter.emit("appointments:fetch");
  });

  // Logout button
  app.addEventListener("click", (e) => {
    if (e.target.id !== "logout") return;

    localStorage.removeItem("auth_token");
    state.auth.token = "";
    window.location.href = "/login";
  });
}
