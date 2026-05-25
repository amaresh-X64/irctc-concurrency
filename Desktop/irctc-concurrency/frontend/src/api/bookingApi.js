import axios from "axios";
import { GIN_URL } from "../constants/constants";

const api = axios.create({ baseURL: GIN_URL });
api.interceptors.request.use((config) => {
  const token = sessionStorage.getItem("irctc_token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});
// ─── Create booking ────────────────────────────
export const createBooking = async (bookingData) => {
  const res = await api.post("/bookings/", bookingData);
  return res.data;
};

// ─── Cancel booking ────────────────────────────
export const cancelBooking = async (bookingId, userId) => {
  const res = await api.delete("/bookings/cancel", {
    data: { booking_id: bookingId, user_id: userId },
  });
  return res.data;
};

// ─── Get bookings by user ──────────────────────
export const getUserBookings = async (userId) => {
  const res = await api.get(`/bookings/user/${userId}`);
  return res.data;
};

// ─── Check seat lock status ────────────────────
export const checkSeatStatus = async (trainId, seatId, journeyDate) => {
  const res = await api.get("/bookings/seat-status", {
    params: { train_id: trainId, seat_id: seatId, journey_date: journeyDate },
  });
  return res.data;
};