const Loader = ({ message = "Loading..." }) => {
  return (
    <div style={{
      display: "flex", flexDirection: "column",
      alignItems: "center", justifyContent: "center",
      padding: "40px", gap: "16px"
    }}>
      <div style={{
        width: "40px", height: "40px",
        border: "4px solid #E5E7EB",
        borderTop: "4px solid #2563EB",
        borderRadius: "50%",
        animation: "spin 0.8s linear infinite"
      }} />
      <p style={{ color: "#6B7280", fontSize: "14px" }}>{message}</p>
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  );
};

export default Loader;