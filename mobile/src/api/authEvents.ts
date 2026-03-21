/** Called by the axios interceptor when token refresh fails. */
let onSessionExpired: (() => void) | null = null

export function setSessionExpiredCallback(cb: () => void) {
  onSessionExpired = cb
}

export function notifySessionExpired() {
  onSessionExpired?.()
}
