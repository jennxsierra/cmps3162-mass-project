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

    f.provider_id = document.querySelector("#filter-provider-id").value.trim();
    f.patient_id = document.querySelector("#filter-patient-id").value.trim();
    f.appt_type_id = document
      .querySelector("#filter-appt-type-id")
      .value.trim();

    f.include_cancelled = document.querySelector(
      "#filter-include-cancelled",
    ).checked;

    f.start_from = document.querySelector("#filter-start-from").value.trim();
    f.start_to = document.querySelector("#filter-start-to").value.trim();

    f.page_size = Number(document.querySelector("#filter-page-size").value);
    f.sort = document.querySelector("#filter-sort").value;

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
