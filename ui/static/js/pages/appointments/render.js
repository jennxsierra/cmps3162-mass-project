import { state } from "../../core/state.js";

// small helper to escape HTML in error messages or text
function esc(s) {
  return String(s ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function renderHeader() {
  return `
    <div class="row space-between">
      <h2>Appointments</h2>
      <button id="logout">Logout</button>
    </div>
  `;
}

function renderLoading() {
  if (!state.appointments.loading) return "";
  return `<div class="loading-box">Loading appointments...</div>`;
}

function renderError() {
  if (!state.appointments.error) return "";
  return `<div class="error-box">${esc(state.appointments.error)}</div>`;
}

function renderFilters() {
  const f = state.appointments.filters;

  return `
    <div id="filters" class="card">
      <div class="grid">
        <label>
          provider_id
          <input id="filter-provider-id" placeholder="e.g. 12" value="${esc(f.provider_id)}" />
        </label>

        <label>
          patient_id
          <input id="filter-patient-id" placeholder="e.g. 99" value="${esc(f.patient_id)}" />
        </label>

        <label>
          appt_type_id
          <input id="filter-appt-type-id" placeholder="e.g. 3" value="${esc(f.appt_type_id)}" />
        </label>

        <label>
          start_from (RFC3339)
          <input id="filter-start-from" placeholder="2026-04-21T00:00:00Z" value="${esc(f.start_from)}" />
        </label>

        <label>
          start_to (RFC3339)
          <input id="filter-start-to" placeholder="2026-04-30T23:59:59Z" value="${esc(f.start_to)}" />
        </label>

        <label>
          page_size
          <select id="filter-page-size">
            ${[5, 10, 20, 50]
              .map(
                (n) => `
              <option value="${n}" ${Number(f.page_size) === n ? "selected" : ""}>${n}</option>
            `,
              )
              .join("")}
          </select>
        </label>

        <label>
          sort
          <select id="filter-sort">
            ${[
              "start_time",
              "-start_time",
              "end_time",
              "-end_time",
              "created_at",
              "-created_at",
              "updated_at",
              "-updated_at",
              "appointment_id",
              "-appointment_id",
              "patient_id",
              "-patient_id",
              "provider_id",
              "-provider_id",
              "appt_type_id",
              "-appt_type_id",
            ]
              .map(
                (s) => `
              <option value="${s}" ${f.sort === s ? "selected" : ""}>${s}</option>
            `,
              )
              .join("")}
          </select>
        </label>

        <label class="inline">
          <input type="checkbox" id="filter-include-cancelled" ${f.include_cancelled ? "checked" : ""} />
          include_cancelled
        </label>
      </div>

      <div style="margin-top: 12px;">
        <button id="apply-filters">Apply</button>
      </div>
    </div>
  `;
}

function renderAppointmentsTable() {
  if (state.appointments.loading || state.appointments.error) return "";

  const rows = state.appointments.data;
  if (!rows.length) return `<p>No appointments found.</p>`;

  return `
    <div class="card">
      <table class="table">
        <thead>
          <tr>
            <th>appointment_id</th>
            <th>start_time</th>
            <th>end_time</th>
            <th>patient_id</th>
            <th>provider_id</th>
            <th>appt_type_id</th>
            <th>reason</th>
            <th>created_at</th>
            <th>updated_at</th>
          </tr>
        </thead>
        <tbody>
          ${rows
            .map(
              (a) => `
            <tr>
              <td>${esc(a.appointment_id)}</td>
              <td>${esc(a.start_time)}</td>
              <td>${esc(a.end_time)}</td>
              <td>${esc(a.patient_id)}</td>
              <td>${esc(a.provider_id)}</td>
              <td>${esc(a.appt_type_id)}</td>
              <td>${esc(a.reason)}</td>
              <td>${esc(a.created_at)}</td>
              <td>${esc(a.updated_at)}</td>
            </tr>
          `,
            )
            .join("")}
        </tbody>
      </table>
    </div>
  `;
}

// Simple prev/next pagination using @metadata
function renderPagination() {
  const m = state.appointments.metadata || {};
  const current = m.current_page;
  const last = m.last_page;

  if (!current || !last) return "";

  return `
    <div id="pagination" class="row center" style="gap: 8px; margin-top: 12px;">
      <button data-page="${Math.max(1, current - 1)}" ${current <= 1 ? "disabled" : ""}>← Prev</button>
      <span>Page ${current} of ${last}</span>
      <button data-page="${Math.min(last, current + 1)}" ${current >= last ? "disabled" : ""}>Next →</button>
    </div>
  `;
}

export function render() {
  const app = document.querySelector("#app");

  app.innerHTML = `
    ${renderHeader()}
    ${renderFilters()}
    ${renderLoading()}
    ${renderError()}
    ${renderAppointmentsTable()}
    ${renderPagination()}
  `;
}
