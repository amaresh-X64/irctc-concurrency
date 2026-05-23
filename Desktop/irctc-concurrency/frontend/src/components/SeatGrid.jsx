import { getSeatColor } from "../helpers/helpers";

const SeatGrid = ({ seats, selectedSeats, onSeatSelect }) => {
  return (
    <div>
      {/* Legend */}
      <div style={{ display: "flex", gap: "16px", marginBottom: "16px", flexWrap: "wrap" }}>
        {[
          { color: "#7C3AED", label: "1st AC" },
          { color: "#2563EB", label: "2nd AC" },
          { color: "#059669", label: "3rd AC" },
          { color: "#D97706", label: "Sleeper" },
          { color: "#6B7280", label: "General" },
          { color: "#EF4444", label: "Booked" },
        ].map((item) => (
          <div key={item.label} style={{ display: "flex", alignItems: "center", gap: "6px" }}>
            <div style={{ width: "14px", height: "14px", borderRadius: "3px", background: item.color }} />
            <span style={{ fontSize: "12px", color: "#6B7280" }}>{item.label}</span>
          </div>
        ))}
      </div>

      {/* Seat grid */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(5, 1fr)", gap: "10px" }}>
        {seats.map((seat) => {
          const color = getSeatColor(seat.seat_type, seat.is_available);
          const isSelected = selectedSeats?.some(s => s.id === seat.id);
          return (
            <button
              key={seat.id}
              onClick={() => seat.is_available && onSeatSelect(seat)}
              disabled={!seat.is_available}
              style={{
                background: isSelected ? "#1E3A5F" : color,
                color: "white",
                border: isSelected ? "3px solid #60A5FA" : "2px solid transparent",
                borderRadius: "8px",
                padding: "10px 6px",
                fontSize: "11px",
                fontWeight: "600",
                cursor: seat.is_available ? "pointer" : "not-allowed",
                opacity: seat.is_available ? 1 : 0.6,
                transition: "all 0.15s",
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                gap: "2px",
              }}
            >
              <span>{seat.seat_number}</span>
              <span style={{ fontSize: "9px", opacity: 0.85 }}>
                {seat.seat_type.replace("_", " ")}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
};

export default SeatGrid;