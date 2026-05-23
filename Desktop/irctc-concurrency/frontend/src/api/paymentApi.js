import axios from "axios";
import { SPRING_URL } from "../constants/constants";

const api = axios.create({ baseURL: SPRING_URL });

// ─── Process payment ───────────────────────────
export const processPayment = async (paymentData) => {
  const res = await api.post("/payments", paymentData);
  return res.data;
};

// ─── Get payment by booking ────────────────────
export const getPaymentByBooking = async (bookingId) => {
  const res = await api.get(`/payments/booking/${bookingId}`);
  return res.data;
};

// ─── Get PNR ───────────────────────────────────
export const getPnr = async (pnrNumber) => {
  const res = await api.get(`/pnr/${pnrNumber}`);
  return res.data;
};