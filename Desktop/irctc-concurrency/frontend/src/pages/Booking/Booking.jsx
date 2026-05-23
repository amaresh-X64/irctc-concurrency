import { useState, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { createBooking } from "../../api/bookingApi";
import SeatGrid from "../../components/SeatGrid";
import Loader from "../../components/Loader";
import { MESSAGES } from "../../constants/constants";
import { formatTime, formatPrice } from "../../helpers/helpers";
import { getSeats, joinWaitlist } from "../../api/trainApi";

const Booking = () => {
  const location = useLocation();
  const navigate  = useNavigate();
  const { train, journeyDate } = location.state || {};

  const [waitlistInfo, setWaitlistInfo]   = useState(null);
  const [seats, setSeats]                 = useState([]);
  const [selectedSeats, setSelectedSeats] = useState([]);
  const [loading, setLoading]             = useState(false);
  const [booking, setBooking]             = useState(false);
  const [error, setError]                 = useState("");

 
  useEffect(() => {
    if (!train) { navigate("/search"); return; }
    fetchSeats();
  }, []);

  useEffect(() => {
    if (!train) return;
    const interval = setInterval(() => {
      fetchSeatsQuietly();
    }, 5000);
    return () => clearInterval(interval); 
  }, [train]);

  const fetchSeats = async () => {
    setLoading(true);
    try {
      const res = await getSeats(train.id, journeyDate);
      setSeats(res.data || []);
    } catch {
      setError(MESSAGES.SERVER_ERROR);
    } finally {
      setLoading(false);
    }
  };


  const fetchSeatsQuietly = async () => {
    try {
      const res = await getSeats(train.id, journeyDate);
      const updatedSeats = res.data || [];
      setSeats(updatedSeats);

      setSelectedSeats(prev =>
        prev.filter(selected =>
          updatedSeats.find(s => s.id === selected.id && s.is_available)
        )
      );
    } catch {
    }
  };

  const handleSeatSelect = (seat) => {
    setSelectedSeats(prev => {
      const exists = prev.some(s => s.id === seat.id);
      if (exists) return prev.filter(s => s.id !== seat.id);
      return [...prev, seat];
    });
  };

  const handleBook = async () => {
    if (selectedSeats.length === 0) return;
    setBooking(true);
    setError("");
    try {
    
      const savedUser = JSON.parse(sessionStorage.getItem("irctc_user"));
      const userId = savedUser?.user_id;

      if (!userId) {
        setError("Please login first!");
        setBooking(false);
        return;
      }

      const bookingResults = [];
      for (const seat of selectedSeats) {
        const res = await createBooking({
          user_id:      userId,
          train_id:     train.id,
          seat_id:      seat.id,
          journey_date: journeyDate,
        });
        if (res.success) {
          bookingResults.push({ booking: res.data, seat });
        } else {
          setError(`Seat ${seat.seat_number} is already taken!`);
          setBooking(false);
          fetchSeats(); 
          return;
        }
      }

      navigate("/payment", {
        state: {
          bookings: bookingResults,
          train,
          journeyDate,
          seats: selectedSeats
        }
      });

    } catch (err) {
      if (err.response?.status === 409) {
        const savedUser = JSON.parse(sessionStorage.getItem("irctc_user"));
        await handleJoinWaitlist(savedUser?.user_id);
      } else {
        setError(MESSAGES.SEAT_TAKEN);
      }
    } finally {
      setBooking(false);
    }
  };

  const handleJoinWaitlist = async (userId) => {
    try {
      const res = await joinWaitlist(userId, train.id, journeyDate);
      if (res.success) setWaitlistInfo(res.data);
    } catch {
      setError("Seat taken. Could not join waitlist.");
    }
  };

  if (!train) return null;

  const totalPrice = selectedSeats.reduce((sum) => sum + train.price, 0);

  return (
    <div style={{ maxWidth: "860px", margin: "0 auto", padding: "32px 16px" }}>

      {/* Train summary */}
      <div style={{
        background: "#1E3A5F", borderRadius: "12px",
        padding: "20px 24px", color: "white", marginBottom: "24px",
        display: "flex", justifyContent: "space-between", flexWrap: "wrap", gap: "12px"
      }}>
        <div>
          <p style={{ fontSize: "18px", fontWeight: "700", margin: "0 0 4px" }}>
            {train.train_name}
          </p>
          <p style={{ opacity: 0.7, fontSize: "13px", margin: 0 }}>
            #{train.train_number} · {journeyDate}
          </p>
        </div>
        <div style={{ textAlign: "right" }}>
          <p style={{ fontSize: "16px", margin: "0 0 4px" }}>
            {formatTime(train.departure_time)} → {formatTime(train.arrival_time)}
          </p>
          <p style={{ fontSize: "18px", fontWeight: "700", margin: 0 }}>
            {formatPrice(train.price)} / seat
          </p>
        </div>
      </div>

      {/* Seat selection */}
      <div style={{
        background: "white", borderRadius: "12px",
        padding: "24px", boxShadow: "0 1px 6px rgba(0,0,0,0.08)",
        marginBottom: "20px"
      }}>
        <h2 style={{ fontSize: "16px", fontWeight: "700", marginBottom: "4px", color: "#111827" }}>
          Select Your Seats
        </h2>
        <p style={{ fontSize: "12px", color: "#6B7280", marginBottom: "20px" }}>
          Click multiple seats to book for family/friends · Seats refresh every 5s
        </p>
        {loading && <Loader message="Loading seats..." />}
        {!loading && (
          <SeatGrid
            seats={seats}
            selectedSeats={selectedSeats}
            onSeatSelect={handleSeatSelect}
          />
        )}
      </div>

      {/* Error */}
      {error && (
        <div style={{
          background: "#FEF2F2", border: "1px solid #FECACA",
          borderRadius: "8px", padding: "14px", color: "#DC2626",
          fontSize: "14px", marginBottom: "16px"
        }}>
          ❌ {error}
        </div>
      )}

      {/* Selected seats summary + book button */}
      {selectedSeats.length > 0 && (
        <div style={{
          background: "#EFF6FF", border: "1px solid #BFDBFE",
          borderRadius: "10px", padding: "16px 20px",
          display: "flex", justifyContent: "space-between",
          alignItems: "center", flexWrap: "wrap", gap: "12px"
        }}>
          <div>
            <p style={{ fontWeight: "600", color: "#1E3A5F", margin: "0 0 4px" }}>
              Selected: {selectedSeats.map(s => s.seat_number).join(", ")}
            </p>
            <p style={{ color: "#6B7280", fontSize: "13px", margin: 0 }}>
              {selectedSeats.length} seat{selectedSeats.length > 1 ? "s" : ""} · Total: {formatPrice(totalPrice)}
            </p>
          </div>
          <button
            onClick={handleBook}
            disabled={booking}
            style={{
              background: booking ? "#9CA3AF" : "#2563EB",
              color: "white", border: "none",
              borderRadius: "8px", padding: "12px 32px",
              fontWeight: "700", fontSize: "15px",
              cursor: booking ? "not-allowed" : "pointer",
            }}
          >
            {booking ? "Booking..." : `Confirm ${selectedSeats.length} Seat${selectedSeats.length > 1 ? "s" : ""} →`}
          </button>
        </div>
      )}

      {/* Waitlist info */}
      {waitlistInfo && (
        <div style={{
          background: "#FEF3C7", border: "1px solid #FCD34D",
          borderRadius: "12px", padding: "20px 24px", marginTop: "16px"
        }}>
          <p style={{ fontWeight: "700", fontSize: "16px", color: "#92400E", margin: "0 0 8px" }}>
            🎫 Seat Taken — You're on the Waitlist!
          </p>
          <p style={{ fontSize: "24px", fontWeight: "700", color: "#78350F", margin: "0 0 4px" }}>
            Position #{waitlistInfo.position}
          </p>
          <p style={{ fontSize: "13px", color: "#92400E", margin: "0 0 12px" }}>
            {waitlistInfo.message}
          </p>
          <button
            onClick={() => navigate("/history")}
            style={{
              marginTop: "8px", background: "#D97706", color: "white",
              border: "none", borderRadius: "8px", padding: "10px 20px",
              fontWeight: "600", fontSize: "14px", cursor: "pointer"
            }}
          >
            View My Waitlist →
          </button>
        </div>
      )}
    </div>
  );
};

export default Booking;