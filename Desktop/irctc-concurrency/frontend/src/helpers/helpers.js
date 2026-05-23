// ─── Format time from HH:MM:SS to HH:MM ───────
export const formatTime = (time) => {
  if (!time) return "";
  return time.substring(0, 5);
};

// ─── Format date to readable ───────────────────
export const formatDate = (date) => {
  if (!date) return "";
  return new Date(date).toLocaleDateString("en-IN", {
    day: "numeric", month: "short", year: "numeric",
  });
};

// ─── Format currency ───────────────────────────
export const formatPrice = (price) => {
  return new Intl.NumberFormat("en-IN", {
    style: "currency", currency: "INR",
  }).format(price);
};

// ─── Generate random user ID for demo ──────────
export const getRandomUserId = () => {
  return Math.floor(Math.random() * 3) + 1;
};

// ─── Get seat color by type ────────────────────
export const getSeatColor = (type, isAvailable) => {
  if (!isAvailable) return "#EF4444";
  const colors = {
    FIRST_AC:  "#7C3AED",
    SECOND_AC: "#2563EB",
    THIRD_AC:  "#059669",
    SLEEPER:   "#D97706",
    GENERAL:   "#6B7280",
  };
  return colors[type] || "#6B7280";
};

// ─── Truncate text ─────────────────────────────
export const truncate = (str, n) => {
  return str?.length > n ? str.substring(0, n) + "..." : str;
};