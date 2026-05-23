import axios from "axios";
import { FASTAPI_URL } from "../constants/constants";

const api = axios.create({ baseURL: FASTAPI_URL });

// ─── Get all trains ────────────────────────────
export const getAllTrains = async (journeyDate = null) => {
  const res = await api.get("/trains/", {
    params: journeyDate ? { journey_date: journeyDate } : {}
  });
  return res.data;
};

// ─── Search trains by route ────────────────────
export const searchTrains = async (source, destination, journeyDate = null) => {
  const res = await api.get("/trains/search", {
    params: {
      source,
      destination,
      ...(journeyDate && { journey_date: journeyDate })
    }
  });
  return res.data;
};

// ─── Get seats for a train ─────────────────────
export const getSeats = async (trainId, journeyDate = null) => {
  const params = journeyDate ? { journey_date: journeyDate } : {};
  const res = await api.get(`/trains/${trainId}/seats`, { params });
  return res.data;
};

// ─── Get waitlist info ─────────────────────────
export const getWaitlistInfo = async (trainId, journeyDate) => {
  const res = await api.get(`/waitlist/${trainId}`, {
    params: { journey_date: journeyDate },
  });
  return res.data;
};

// ─── Join waitlist ─────────────────────────────
export const joinWaitlist = async (userId, trainId, journeyDate) => {
  const res = await api.post("/waitlist/join", null, {
    params: {
      user_id:      userId,
      train_id:     trainId,
      journey_date: journeyDate
    }
  });
  return res.data;
};

// ─── Get user waitlist ─────────────────────────
export const getUserWaitlist = async (userId) => {
  const res = await api.get(`/waitlist/user/${userId}`);
  return res.data;
};