import { useNavigate } from "react-router-dom";
import { formatTime, formatPrice } from "../helpers/helpers";
import { joinWaitlist } from "../api/trainApi";
import { useState } from "react";

const TrainCard = ({ train, journeyDate }) => {
  const navigate = useNavigate();
  const [joining, setJoining]           = useState(false);
  const [joined, setJoined]             = useState(false);
  const [waitlistPos, setWaitlistPos]   = useState(null);
  const [showCount, setShowCount]       = useState(false);
  const [seatCount, setSeatCount]       = useState(1);
  const [waitlistDate, setWaitlistDate] = useState(
    journeyDate || new Date().toISOString().split("T")[0]

  );


  // This ensures correct user even across tabs
  const getUserId = () => {
    const savedUser = JSON.parse(sessionStorage.getItem("irctc_user"));
    return savedUser?.user_id || null;
  };

  const handleWaitlist = async () => {
    const userId = getUserId();
    if (!userId) {
      alert("Please login first!");
      return;
    }

    setJoining(true);
    try {
      let lastPos = null;
      for (let i = 0; i < seatCount; i++) {
        const res = await joinWaitlist(userId, train.id, waitlistDate);
        if (res.success) lastPos = res.data.position;
      }
      setJoined(true);
      setWaitlistPos(lastPos);
      setShowCount(false);
    } catch {
      alert("Failed to join waitlist. Please try again.");
    } finally {
      setJoining(false);
    }
  };

  return (
    <div style={{
      border: "1px solid #E5E7EB",
      borderRadius: "12px",
      padding: "20px 24px",
      background: "white",
      boxShadow: "0 1px 4px rgba(0,0,0,0.06)",
      display: "flex",
      alignItems: "center",
      justifyContent: "space-between",
      gap: "16px",
      flexWrap: "wrap",
    }}>
      {/* Train info */}
      <div>
        <p style={{ fontWeight: "700", fontSize: "16px", color: "#111827", margin: "0 0 4px" }}>
          {train.train_name}
        </p>
        <p style={{ color: "#6B7280", fontSize: "13px", margin: 0 }}>
          #{train.train_number}
        </p>
      </div>

      {/* Route + time */}
      <div style={{ textAlign: "center" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
          <div>
            <p style={{ fontWeight: "700", fontSize: "18px", margin: 0 }}>
              {formatTime(train.departure_time)}
            </p>
            <p style={{ color: "#6B7280", fontSize: "12px", margin: 0 }}>
              {train.source}
            </p>
          </div>
          <div style={{ color: "#9CA3AF", fontSize: "20px" }}>→</div>
          <div>
            <p style={{ fontWeight: "700", fontSize: "18px", margin: 0 }}>
              {formatTime(train.arrival_time)}
            </p>
            <p style={{ color: "#6B7280", fontSize: "12px", margin: 0 }}>
              {train.destination}
            </p>
          </div>
        </div>
      </div>

      {/* Seats + price */}
      <div style={{ textAlign: "center" }}>
        <p style={{
          color: train.available_seats > 0 ? "#059669" : "#EF4444",
          fontWeight: "600", fontSize: "14px", margin: "0 0 4px"
        }}>
          {train.available_seats > 0
            ? `${train.available_seats} seats available`
            : "Fully booked"}
        </p>
        <p style={{ fontWeight: "700", fontSize: "18px", color: "#1E3A5F", margin: 0 }}>
          {formatPrice(train.price)}
        </p>
      </div>

      {/* Book / Waitlist button */}
      <div style={{ textAlign: "center" }}>
        {train.available_seats > 0 ? (
          <button
            onClick={() => navigate("/booking", { state: { train, journeyDate } })}
            style={{
              background: "#2563EB", color: "white",
              border: "none", borderRadius: "8px",
              padding: "10px 24px", fontWeight: "600",
              fontSize: "14px", cursor: "pointer",
            }}
          >
            Book Now
          </button>
        ) : joined ? (
          <div style={{ textAlign: "center" }}>
            <p style={{
              background: "#FEF3C7", color: "#92400E",
              border: "1px solid #FCD34D",
              borderRadius: "8px", padding: "8px 16px",
              fontWeight: "600", fontSize: "13px", margin: "0 0 4px"
            }}>
              ✅ Waitlisted {seatCount > 1 ? `(${seatCount} seats)` : ""} — Position #{waitlistPos}
            </p>
            <button
              onClick={() => navigate("/history")}
              style={{
                background: "transparent", color: "#2563EB",
                border: "none", fontSize: "12px",
                cursor: "pointer", textDecoration: "underline"
              }}
            >
              View Waitlist →
            </button>
          </div>
        ) : showCount ? (
          <div style={{
            display: "flex", flexDirection: "column",
            gap: "8px", alignItems: "center", minWidth: "200px"
          }}>
            {/* Date picker */}
            <p style={{ fontSize: "12px", color: "#374151", fontWeight: "600", margin: 0 }}>
              Journey Date
            </p>
            <input
              type="date"
              value={waitlistDate}
              min={new Date().toISOString().split("T")[0]}
              onChange={(e) => setWaitlistDate(e.target.value)}
              style={{
                padding: "6px 10px", borderRadius: "8px",
                border: "1px solid #D1D5DB", fontSize: "13px",
                width: "100%"
              }}
            />

            {/* Seat count */}
            <p style={{ fontSize: "12px", color: "#374151", fontWeight: "600", margin: 0 }}>
              How many seats?
            </p>
            <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              <button
                onClick={() => setSeatCount(c => Math.max(1, c - 1))}
                style={{
                  width: "28px", height: "28px", borderRadius: "50%",
                  border: "1px solid #D1D5DB", background: "white",
                  cursor: "pointer", fontWeight: "700", fontSize: "16px"
                }}
              >−</button>
              <span style={{ fontWeight: "700", fontSize: "16px", minWidth: "20px", textAlign: "center" }}>
                {seatCount}
              </span>
              <button
                onClick={() => setSeatCount(c => Math.min(6, c + 1))}
                style={{
                  width: "28px", height: "28px", borderRadius: "50%",
                  border: "1px solid #D1D5DB", background: "white",
                  cursor: "pointer", fontWeight: "700", fontSize: "16px"
                }}
              >+</button>
            </div>

            {/* Action buttons */}
            <div style={{ display: "flex", gap: "8px" }}>
              <button
                onClick={() => setShowCount(false)}
                style={{
                  background: "white", color: "#6B7280",
                  border: "1px solid #D1D5DB", borderRadius: "8px",
                  padding: "6px 14px", fontSize: "12px", cursor: "pointer"
                }}
              >
                Cancel
              </button>
              <button
                onClick={handleWaitlist}
                disabled={joining}
                style={{
                  background: joining ? "#9CA3AF" : "#D97706",
                  color: "white", border: "none",
                  borderRadius: "8px", padding: "6px 14px",
                  fontWeight: "600", fontSize: "12px",
                  cursor: joining ? "not-allowed" : "pointer"
                }}
              >
                {joining ? "Joining..." : "Confirm"}
              </button>
            </div>
          </div>
        ) : (
          <button
            onClick={() => setShowCount(true)}
            style={{
              background: "#D97706", color: "white",
              border: "none", borderRadius: "8px",
              padding: "10px 24px", fontWeight: "600",
              fontSize: "14px", cursor: "pointer",
            }}
          >
            ⏳ Join Waitlist
          </button>
        )}
      </div>
    </div>
  );
};

export default TrainCard;