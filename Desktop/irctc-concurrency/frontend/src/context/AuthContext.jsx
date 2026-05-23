import { createContext, useContext, useState, useEffect } from "react";

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser]     = useState(null);
  const [token, setToken]   = useState(null);
  const [ready, setReady]   = useState(false);

  // ─── Load from sessionStorage once on startup ───
  // ✅ sessionStorage is isolated per tab — Nive and Nisha won't mix!
  useEffect(() => {
    try {
      const savedUser  = sessionStorage.getItem("irctc_user");
      const savedToken = sessionStorage.getItem("irctc_token");
      if (savedUser && savedToken) {
        const parsedUser = JSON.parse(savedUser);
        setUser(parsedUser);
        setToken(savedToken);
        window.__auth_token__ = savedToken;
      }
    } catch {
      sessionStorage.removeItem("irctc_user");
      sessionStorage.removeItem("irctc_token");
    } finally {
      setReady(true);
    }
  }, []);

  const saveAuth = (userData, accessToken) => {
    setUser(userData);
    setToken(accessToken);
    sessionStorage.setItem("irctc_user", JSON.stringify(userData));
    sessionStorage.setItem("irctc_token", accessToken);
    window.__auth_token__ = accessToken;
  };

  const logout = () => {
    setUser(null);
    setToken(null);
    sessionStorage.removeItem("irctc_user");
    sessionStorage.removeItem("irctc_token");
    window.__auth_token__ = null;
  };

  // ─── Don't render until auth is loaded ────────
  if (!ready) return null;

  return (
    <AuthContext.Provider value={{ user, token, saveAuth, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);