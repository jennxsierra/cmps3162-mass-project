// state object to hold the application state
export const state = {
  auth: {
    token: localStorage.getItem("auth_token") || "",
    loading: false,
    error: null,
  },

  // appointments entity
  appointments: {
    data: [],
    metadata: {},

    filters: {
      page: 1,
      page_size: 5,
      sort: "start_time",
      provider_id: "",
      patient_id: "",
      appt_type_id: "",
      include_cancelled: true,
      start_from: "",
      start_to: "",
    },

    loading: false,
    error: null,
  },
};
