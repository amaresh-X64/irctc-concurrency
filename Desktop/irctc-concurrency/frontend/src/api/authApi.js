import axios from "axios";
import { FASTAPI_URL } from "../constants/constants";

const api = axios.create({ baseURL: FASTAPI_URL });

// ─── Register ──────────────────────────────────
export const register = async (data) => {
  const res = await api.post("/auth/register", data);
  return res.data;
};

// ─── Login ─────────────────────────────────────
export const login = async (data) => {
  const res = await api.post("/auth/login", data);
  return res.data;
};