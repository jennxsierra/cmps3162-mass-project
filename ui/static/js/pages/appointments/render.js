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

function formatDate(d) {
  if (!d) return "";
  const date = new Date(d);
  if (isNaN(date.getTime())) return esc(d);
  return date.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  });
}

function renderHeader() {
  return `
    <div class="row space-between" style="margin-bottom: 24px; padding: 0 8px;">
      <h2>Appointments</h2>
      <button id="logout" class="secondary">Logout</button>
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
          Page Size
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
          Sort By
          <select id="filter-sort">
            ${[
              { val: "start_time", label: "Start Time (Asc)" },
              { val: "-start_time", label: "Start Time (Desc)" },
              { val: "end_time", label: "End Time (Asc)" },
              { val: "-end_time", label: "End Time (Desc)" },
              { val: "created_at", label: "Created At (Asc)" },
              { val: "-created_at", label: "Created At (Desc)" },
              { val: "updated_at", label: "Updated At (Asc)" },
              { val: "-updated_at", label: "Updated At (Desc)" },
              { val: "appointment_id", label: "ID (Asc)" },
              { val: "-appointment_id", label: "ID (Desc)" },
            ]
              .map(
                (s) => `
              <option value="${s.val}" ${f.sort === s.val ? "selected" : ""}>${s.label}</option>
            `,
              )
              .join("")}
          </select>
        </label>

        <label class="inline">
          <input type="checkbox" id="filter-include-cancelled" ${f.include_cancelled ? "checked" : ""} />
          Include Cancelled
        </label>
      </div>

      <div style="margin-top: 24px; display: flex; justify-content: flex-end;">
        <button id="apply-filters">Apply Filters</button>
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
      <div class="table-container">
        <table class="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Start Time</th>
              <th>End Time</th>
              <th>Patient</th>
              <th>Provider</th>
              <th>Type</th>
              <th>Reason</th>
              <th>Created At</th>
              <th>Updated At</th>
            </tr>
          </thead>
          <tbody>
            ${rows
              .map(
                (a) => `
              <tr>
                <td>${esc(a.appointment_id)}</td>
                <td>${formatDate(a.start_time)}</td>
                <td>${formatDate(a.end_time)}</td>
                <td>${esc(a.patient_name || a.patient_id)}</td>
                <td>${esc(a.provider_name || a.provider_id)}</td>
                <td>${esc(a.appt_type_name || a.appt_type_id)}</td>
                <td>${esc(a.reason)}</td>
                <td>${formatDate(a.created_at)}</td>
                <td>${formatDate(a.updated_at)}</td>
              </tr>
            `,
              )
              .join("")}
          </tbody>
        </table>
      </div>
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
    <div id="pagination" class="row center" style="gap: 12px; margin-top: 24px;">
      <button class="secondary" data-page="${Math.max(1, current - 1)}" ${current <= 1 ? "disabled" : ""}>← Prev</button>
      <span style="font-weight: 500; color: var(--color-green-darker);">Page ${current} of ${last}</span>
      <button class="secondary" data-page="${Math.min(last, current + 1)}" ${current >= last ? "disabled" : ""}>Next →</button>
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
