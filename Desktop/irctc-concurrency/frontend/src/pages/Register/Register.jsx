import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { register } from "../../api/authApi";
import { useAuth } from "../../context/AuthContext";

const Register = () => {
  const navigate = useNavigate();
  const { saveAuth } = useAuth();

  const [form, setForm]       = useState({ name: "", email: "", password: "", phone: "" });
  const [loading, setLoading] = useState(false);
  const [error, setError]     = useState("");

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleRegister = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await register(form);
      if (res.success) {
        saveAuth(res.data, res.data.access_token);
        navigate("/search");
      } else {
        setError(res.message);
      }
    } catch {
      setError("Registration failed. Please try again.");
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
            Create Account
          </h1>
          <p style={{ color: "#6B7280", fontSize: "14px", margin: 0 }}>
            Join IRCTC Clone today
          </p>
        </div>

        <form onSubmit={handleRegister}>
          {[
            { label: "FULL NAME", name: "name",     type: "text",     placeholder: "Your name" },
            { label: "EMAIL",     name: "email",    type: "email",    placeholder: "you@example.com" },
            { label: "PASSWORD",  name: "password", type: "password", placeholder: "••••••••" },
            { label: "PHONE",     name: "phone",    type: "text",     placeholder: "9999999999" },
          ].map((field) => (
            <div key={field.name} style={{ marginBottom: "16px" }}>
              <label style={{ fontSize: "12px", fontWeight: "600", color: "#374151", display: "block", marginBottom: "6px" }}>
                {field.label}
              </label>
              <input
                type={field.type} name={field.name}
                required={field.name !== "phone"}
                style={inputStyle}
                placeholder={field.placeholder}
                value={form[field.name]}
                onChange={handleChange}
              />
            </div>
          ))}

          {error && (
            <div style={{
              background: "#FEF2F2", border: "1px solid #FECACA",
              borderRadius: "8px", padding: "12px",
              color: "#DC2626", fontSize: "13px", marginBottom: "16px"
            }}>
              ❌ {error}
            </div>
          )}

          <button type="submit" disabled={loading} style={{
            width: "100%", background: loading ? "#9CA3AF" : "#2563EB",
            color: "white", border: "none", borderRadius: "10px",
            padding: "14px", fontWeight: "700", fontSize: "15px",
            cursor: loading ? "not-allowed" : "pointer",
          }}>
            {loading ? "Creating account..." : "Register"}
          </button>
        </form>

        <p style={{ textAlign: "center", marginTop: "20px", fontSize: "14px", color: "#6B7280" }}>
          Already have an account?{" "}
          <Link to="/login" style={{ color: "#2563EB", fontWeight: "600" }}>
            Login
          </Link>
        </p>
      </div>
    </div>
  );
};

export default Register;