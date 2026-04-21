import { emitter } from "./event-emitter.js";
import { state } from "./state.js";

const API_BASE = "/v1";

// DataService provides methods to interact with the backend API
// for authentication and fetching appointments.
export const DataService = {
  async login({ email, password }) {
    try {
      const res = await fetch(`${API_BASE}/tokens/authentication`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      if (!res.ok)
        throw new Error(`Login failed: ${res.status} ${res.statusText}`);

      const data = await res.json();

      // extract the token from the response
      const token = data?.authentication_token?.token;
      if (!token) throw new Error("Token missing from response.");

      emitter.emit("auth:loginSuccess", token);
    } catch (err) {
      emitter.emit("auth:loginError", err.message);
    }
  },

  async fetchAppointments() {
    try {
      const params = new URLSearchParams();
      for (const [k, v] of Object.entries(state.appointments.filters)) {
        if (v === "" || v === null || typeof v === "undefined") continue;
        params.set(k, String(v));
      }

      const res = await fetch(`${API_BASE}/appointments?${params}`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${state.auth.token}`,
        },
      });

      if (!res.ok)
        throw new Error(`Server Error: ${res.status} ${res.statusText}`);

      const payload = await res.json();
      emitter.emit("appointments:fetched", payload);
    } catch (err) {
      emitter.emit("appointments:error", err.message);
    }
  },
};
