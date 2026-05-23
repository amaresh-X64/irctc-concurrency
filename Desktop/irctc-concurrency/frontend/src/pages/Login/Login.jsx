import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { login } from "../../api/authApi";
import { useAuth } from "../../context/AuthContext";

const Login = () => {
  const navigate = useNavigate();
  const { saveAuth } = useAuth();

  const [email, setEmail]       = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading]   = useState(false);
  const [error, setError]       = useState("");
  

  const handleLogin = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await login({ email, password });
      if (res.success) {
      saveAuth(res.data, res.data.access_token);
      setAuth(                                    
    { id: res.data.user_id, 
      name: res.data.name, 
      email: res.data.email 
    },
    res.data.access_token
  );
  navigate("/search");
}

      else {
        setError(res.message);
      }
    } catch {
      setError("Invalid email or password");
    } finally {
      setLoading(false);
    }
  };

  const inputStyle = {
    width: "100%", padding: "12px 14px",
    borderRadius: "8px", border: "1px solid #D1D5DB",
    fontSize: "14px", outline: "none", boxSizing: "border-box",
  };

  return (
    <div style={{
      minHeight: "100vh", display: "flex",
      alignItems: "center", justifyContent: "center",
      background: "#F3F4F6"
    }}>
      <div style={{
        background: "white", borderRadius: "16px",
        padding: "40px", width: "100%", maxWidth: "400px",
        boxShadow: "0 4px 20px rgba(0,0,0,0.1)"
      }}>
        <div style={{ textAlign: "center", marginBottom: "32px" }}>
          <div style={{ fontSize: "40px", marginBottom: "8px" }}>🚂</div>
          <h1 style={{ fontSize: "22px", fontWeight: "700", margin: "0 0 4px" }}>
            Welcome Back
          </h1>
          <p style={{ color: "#6B7280", fontSize: "14px", margin: 0 }}>
            Login to book your train tickets
          </p>
        </div>

        <form onSubmit={handleLogin}>
          <div style={{ marginBottom: "16px" }}>
            <label style={{ fontSize: "12px", fontWeight: "600", color: "#374151", display: "block", marginBottom: "6px" }}>
              EMAIL
            </label>
            <input
              type="email" required style={inputStyle}
              placeholder="you@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
          <div style={{ marginBottom: "24px" }}>
            <label style={{ fontSize: "12px", fontWeight: "600", color: "#374151", display: "block", marginBottom: "6px" }}>
              PASSWORD
            </label>
            <input
              type="password" required style={inputStyle}
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>

          {error && (
            <div style={{
              background: "#FEF2F2", border: "1px solid #FECACA",
              borderRadius: "8px", padding: "12px",
              color: "#DC2626", fontSize: "13px", marginBottom: "16px"
            }}>
               {error}
            </div>
          )}

          <button type="submit" disabled={loading} style={{
            width: "100%", background: loading ? "#9CA3AF" : "#2563EB",
            color: "white", border: "none", borderRadius: "10px",
            padding: "14px", fontWeight: "700", fontSize: "15px",
            cursor: loading ? "not-allowed" : "pointer",
          }}>
            {loading ? "Logging in..." : "Login"}
          </button>
        </form>

        <p style={{ textAlign: "center", marginTop: "20px", fontSize: "14px", color: "#6B7280" }}>
          Don't have an account?{" "}
          <Link to="/register" style={{ color: "#2563EB", fontWeight: "600" }}>
            Register
          </Link>
        </p>
      </div>
    </div>
  );
};

export default Login;