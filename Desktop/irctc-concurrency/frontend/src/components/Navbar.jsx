import { Link, useLocation } from "react-router-dom";
import { ROUTES } from "../constants/constants";

const Navbar = () => {
  const location = useLocation();

  const navStyle = {
    background: "#1E3A5F",
    padding: "0 32px",
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    height: "60px",
    boxShadow: "0 2px 8px rgba(0,0,0,0.2)",
    position: "sticky", top: 0, zIndex: 100,
  };

  const linkStyle = (path) => ({
    color: location.pathname === path ? "#60A5FA" : "#CBD5E1",
    textDecoration: "none",
    fontSize: "14px",
    fontWeight: location.pathname === path ? "600" : "400",
    padding: "6px 12px",
    borderRadius: "6px",
    background: location.pathname === path ? "rgba(96,165,250,0.1)" : "transparent",
  });

  return (
    <nav style={navStyle}>
      {/* Logo */}
      <Link to={ROUTES.HOME} style={{
        color: "white", textDecoration: "none",
        fontSize: "20px", fontWeight: "700",
        display: "flex", alignItems: "center", gap: "8px"
      }}>
        🚂 IRCTC Clone
      </Link>

      {/* Links */}
      <div style={{ display: "flex", gap: "8px", alignItems: "center" }}>
        <Link to={ROUTES.SEARCH} style={linkStyle(ROUTES.SEARCH)}>
          Search Trains
        </Link>
        <Link to={ROUTES.DEMO} style={{
          ...linkStyle(ROUTES.DEMO),
          background: location.pathname === ROUTES.DEMO
            ? "rgba(239,68,68,0.2)" : "rgba(239,68,68,0.1)",
          color: "#FCA5A5",
          border: "1px solid rgba(239,68,68,0.3)",
        }}>
          ⚡ Concurrency Demo
        </Link>
        <Link to={ROUTES.HISTORY} style={linkStyle(ROUTES.HISTORY)}>
            🎫 Booking History
        </Link>
      </div>
    </nav>
  );
};

export default Navbar;