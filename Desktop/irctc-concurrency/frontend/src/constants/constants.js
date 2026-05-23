// ─── Service URLs ──────────────────────────────
export const FASTAPI_URL = "http://localhost:8001/api/v1";
export const GIN_URL     = "http://localhost:8002/api/v1";
export const SPRING_URL  = "http://localhost:8003/api/v1";

// ─── Routes ────────────────────────────────────
export const ROUTES = {
  HOME:        "/",
  SEARCH:      "/search",
  BOOKING:     "/booking",
  PAYMENT:     "/payment",
  CONFIRM:     "/confirmation",
  HISTORY: "/history",
  DEMO:        "/demo",
};

// ─── Payment Methods ───────────────────────────
export const PAYMENT_METHODS = [
  { value: "MOCK_UPI",        label: "UPI" },
  { value: "MOCK_CARD",       label: "Credit / Debit Card" },
  { value: "MOCK_NETBANKING", label: "Net Banking" },
];

// ─── Seat Types ────────────────────────────────
export const SEAT_TYPE_COLORS = {
  FIRST_AC:  "#7C3AED",
  SECOND_AC: "#2563EB",
  THIRD_AC:  "#059669",
  SLEEPER:   "#D97706",
  GENERAL:   "#6B7280",
};

// ─── Messages ──────────────────────────────────
export const MESSAGES = {
  SEAT_TAKEN:      "Seat already taken! Try another seat.",
  BOOKING_SUCCESS: "Seat booked successfully!",
  PAYMENT_SUCCESS: "Payment successful! Your PNR is ready.",
  PAYMENT_FAILED:  "Payment failed. Please try again.",
  NO_TRAINS:       "No trains found for this route.",
  SERVER_ERROR:    "Something went wrong. Please try again.",
};