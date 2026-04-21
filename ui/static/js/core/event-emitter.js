// createEmitter implements an emitter as part of the Observer pattern API.
// It uses a Map to store event listeners linked to an array of callback functions.
function createEmitter() {
  const listeners = new Map();

  // on registers a callback function for a specific event.
  function on(event, callback) {
    // if the event doesn't exist, then create it within the listeners Map.
    if (!listeners.has(event)) listeners.set(event, []);
    listeners.get(event).push(callback);
  }

  // emit triggers all callback functions associated with a specific event,
  // passing the provided data to each callback.
  function emit(event, data) {
    if (!listeners.has(event)) return;
    listeners.get(event).forEach((cb) => cb(data));
  }

  // off removes a callback function from a specific event.
  function off(event, callback) {
    if (!listeners.has(event)) return;
    listeners.set(
      event,
      listeners.get(event).filter((cb) => cb !== callback),
    );
  }

  return { on, emit, off };
}

export const emitter = createEmitter();
