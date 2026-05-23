import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { getUserWaitlist } from "../../api/trainApi";
import Loader from "../../components/Loader";
import { MESSAGES } from "../../constants/constants";
import { getUserBookings, cancelBooking } from "../../api/bookingApi";

const STATUS_COLORS = {
  CONFIRMED:  { bg: "#D1FAE5", color: "#065F46", border: "#6EE7B7" },
  PENDING:    { bg: "#FEF3C7", color: "#92400E", border: "#FCD34D" },
  CANCELLED:  { bg: "#FEE2E2", color: "#991B1B", border: "#FCA5A5" },
  WAITLISTED: { bg: "#EDE9FE", color: "#5B21B6", border: "#C4B5FD" },
};

const TRAIN_NAMES = {
  1:  "Chennai Express",
  2:  "Rajdhani Express",
  3:  "Shatabdi Express",
  4:  "Duronto Express",
  5:  "Vande Bharat Express",
  6:  "Garib Rath Express",
  7:  "Tejas Express",
  8:  "Intercity Express",
  9:  "Konkan Express",
  10: "Deccan Queen",
  11: "Mysore Express",
  12: "Coromandel Express",
  13: "Humsafar Express",
  14: "Jan Shatabdi",
  15: "Udyan Express",
};

const BookingHistory = () => {
  const navigate = useNavigate();

  const currentUser = JSON.parse(sessionStorage.getItem("irctc_user"));
  const userId = currentUser?.user_id;

  const [bookings, setBookings]     = useState([]);
  const [loading, setLoading]       = useState(false);
  const [error, setError]           = useState("");
  const [activeTab, setActiveTab]   = useState("bookings");
  const [waitlist, setWaitlist]     = useState([]);
  const [wLoading, setWLoading]     = useState(false);
  const [cancelling, setCancelling] = useState(null);

  useEffect(() => {
    fetchBookings();
    fetchWaitlist();
  }, []);

 
  const fetchBookings = async () => {
    setLoading(true);
    setError("");
    try {
      const res = await getUserBookings(userId); 
      if (res.success) {
        setBookings(res.data || []);
      } else {
        setError(MESSAGES.SERVER_ERROR);
      }
    } catch {
      setError(MESSAGES.SERVER_ERROR);
    } finally {
      setLoading(false);
    }
  };

  const fetchWaitlist = async () => {
    setWLoading(true);
    try {
      const res = await getUserWaitlist(userId); 
      if (res.success) setWaitlist(res.data || []);
    } catch {
      console.error("Waitlist fetch failed");
    } finally {
      setWLoading(false);
    }
  };

  const handleCancel = async (bookingId) => {
    if (!window.confirm("Are you sure you want to cancel this booking?")) return;
    setCancelling(bookingId);
    try {
      const res = await cancelBooking(bookingId, userId); // ✅ real user id
      if (res.success) {
        setBookings(prev => prev.map(b =>
          b.id === bookingId ? { ...b, status: "CANCELLED" } : b
        ));
        fetchBookings();
        fetchWaitlist();
      }
    } catch {
      alert("Failed to cancel booking. Please try again.");
    } finally {
      setCancelling(null);
    }
  };

  const formatDate = (dateStr) => {
    return new Date(dateStr).toLocaleDateString("en-IN", {
      day: "2-digit", month: "short", year: "numeric"
    });
  };

  const formatDateTime = (dateStr) => {
    return new Date(dateStr).toLocaleString("en-IN", {
      day: "2-digit", month: "short", year: "numeric",
      hour: "2-digit", minute: "2-digit"
    });
  };

  return (
    <div style={{ maxWidth: "860px", margin: "0 auto", padding: "32px 16px" }}>

      {/* Header */}
      <div style={{
        display: "flex", justifyContent: "space-between",
        alignItems: "center", marginBottom: "24px"
      }}>
        <div>
          <h1 style={{ fontSize: "24px", fontWeight: "700", color: "#111827", margin: "0 0 4px" }}>
            🎫 Booking History
          </h1>
          <p style={{ color: "#6B7280", margin: 0 }}>
            All your past and upcoming train bookings
          </p>
        </div>
        <button
          onClick={() => navigate("/search")}
          style={{
            background: "#2563EB", color: "white",
            border: "none", borderRadius: "8px",
            padding: "10px 20px", fontWeight: "600",
            fontSize: "14px", cursor: "pointer"
          }}
        >
          + Book New Ticket
        </button>
      </div>

      {/* Tabs */}
      <div style={{
        display: "flex", gap: "4px",
        background: "#F3F4F6", borderRadius: "10px",
        padding: "4px", marginBottom: "24px", width: "fit-content"
      }}>
        {["bookings", "waitlist"].map(tab => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            style={{
              padding: "8px 20px", borderRadius: "8px",
              border: "none", fontWeight: "600", fontSize: "13px",
              cursor: "pointer",
              background: activeTab === tab ? "white" : "transparent",
              color: activeTab === tab ? "#1E3A5F" : "#6B7280",
              boxShadow: activeTab === tab ? "0 1px 4px rgba(0,0,0,0.1)" : "none"
            }}
          >
            {tab === "bookings"
              ? `🎫 Bookings (${bookings.length})`
              : `⏳ Waitlist (${waitlist.length})`}
          </button>
        ))}
      </div>

      {/* BOOKINGS TAB */}
      {activeTab === "bookings" && (
        <div>
          {loading && <Loader message="Fetching your bookings..." />}

          {error && (
            <div style={{
              background: "#FEF2F2", border: "1px solid #FECACA",
              borderRadius: "8px", padding: "16px",
              color: "#DC2626", fontSize: "14px", marginBottom: "16px"
            }}>
              {error}
            </div>
          )}

          {!loading && !error && bookings.length === 0 && (
            <div style={{
              background: "white", borderRadius: "12px",
              padding: "48px", textAlign: "center",
              boxShadow: "0 1px 6px rgba(0,0,0,0.08)"
            }}>
              <p style={{ fontSize: "48px", margin: "0 0 16px" }}>🎫</p>
              <p style={{ fontSize: "18px", fontWeight: "600", color: "#111827", margin: "0 0 8px" }}>
                No bookings yet
              </p>
              <p style={{ color: "#6B7280", margin: "0 0 24px" }}>
                Search for trains and book your first ticket!
              </p>
              <button
                onClick={() => navigate("/search")}
                style={{
                  background: "#2563EB", color: "white",
                  border: "none", borderRadius: "8px",
                  padding: "12px 28px", fontWeight: "600",
                  fontSize: "14px", cursor: "pointer"
                }}
              >
                Search Trains
              </button>
            </div>
          )}

          {!loading && bookings.length > 0 && (
            <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
              {bookings.map((booking) => {
                const statusStyle = STATUS_COLORS[booking.status] || STATUS_COLORS.PENDING;
                return (
                  <div key={booking.id} style={{
                    background: "white", borderRadius: "12px",
                    padding: "20px 24px", boxShadow: "0 1px 6px rgba(0,0,0,0.08)",
                    border: "1px solid #F3F4F6"
                  }}>
                    {/* Top row */}
                    <div style={{
                      display: "flex", justifyContent: "space-between",
                      alignItems: "flex-start", marginBottom: "16px"
                    }}>
                      <div>
                        <p style={{ fontSize: "17px", fontWeight: "700", color: "#111827", margin: "0 0 4px" }}>
                          {TRAIN_NAMES[booking.train_id] || `Train #${booking.train_id}`}
                        </p>
                        <p style={{ fontSize: "13px", color: "#6B7280", margin: 0 }}>
                          Booking ID: #{booking.id} · Seat: {booking.seat_number || `ID: ${booking.seat_id}`}
                        </p>
                      </div>
                      <div style={{ display: "flex", gap: "8px", alignItems: "center" }}>
                        <span style={{
                          background: statusStyle.bg,
                          color: statusStyle.color,
                          border: `1px solid ${statusStyle.border}`,
                          borderRadius: "20px", padding: "4px 12px",
                          fontSize: "12px", fontWeight: "700",
                        }}>
                          {booking.status}
                        </span>
                        {booking.status === "CONFIRMED" && (
                          <button
                            onClick={() => handleCancel(booking.id)}
                            disabled={cancelling === booking.id}
                            style={{
                              background: cancelling === booking.id ? "#9CA3AF" : "#FEE2E2",
                              color: cancelling === booking.id ? "white" : "#DC2626",
                              border: "1px solid #FECACA",
                              borderRadius: "20px", padding: "4px 12px",
                              fontSize: "12px", fontWeight: "700",
                              cursor: cancelling === booking.id ? "not-allowed" : "pointer"
                            }}
                          >
                            {cancelling === booking.id ? "Cancelling..." : "❌ Cancel"}
                          </button>
                        )}
                      </div>
                    </div>

                    {/* Details row */}
                    <div style={{
                      display: "grid", gridTemplateColumns: "1fr 1fr",
                      gap: "12px", paddingTop: "16px",
                      borderTop: "1px solid #F3F4F6"
                    }}>
                      <div>
                        <p style={{
                          fontSize: "11px", color: "#9CA3AF",
                          fontWeight: "600", margin: "0 0 4px",
                          textTransform: "uppercase"
                        }}>
                          Journey Date
                        </p>
                        <p style={{ fontSize: "14px", fontWeight: "600", color: "#111827", margin: 0 }}>
                          {formatDate(booking.journey_date)}
                        </p>
                      </div>
                      <div>
                        <p style={{
                          fontSize: "11px", color: "#9CA3AF",
                          fontWeight: "600", margin: "0 0 4px",
                          textTransform: "uppercase"
                        }}>
                          Booked At
                        </p>
                        <p style={{ fontSize: "14px", fontWeight: "600", color: "#111827", margin: 0 }}>
                          {formatDateTime(booking.booked_at)}
                        </p>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}

      {/* WAITLIST TAB */}
      {activeTab === "waitlist" && (
        <div>
          {wLoading && <Loader message="Fetching waitlist..." />}

          {!wLoading && waitlist.length === 0 && (
            <div style={{
              background: "white", borderRadius: "12px",
              padding: "48px", textAlign: "center",
              boxShadow: "0 1px 6px rgba(0,0,0,0.08)"
            }}>
              <p style={{ fontSize: "48px", margin: "0 0 16px" }}>⏳</p>
              <p style={{ fontSize: "18px", fontWeight: "600", color: "#111827", margin: "0 0 8px" }}>
                No waitlist entries
              </p>
              <p style={{ color: "#6B7280", margin: 0 }}>
                When seats are full, you'll be added to the waitlist automatically.
              </p>
            </div>
          )}

          {!wLoading && waitlist.length > 0 && (
            <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
              {waitlist.map((w) => (
                <div key={w.id} style={{
                  background: "white", borderRadius: "12px",
                  padding: "20px 24px", boxShadow: "0 1px 6px rgba(0,0,0,0.08)",
                  border: "1px solid #FCD34D"
                }}>
                  <div style={{
                    display: "flex", justifyContent: "space-between",
                    alignItems: "flex-start"
                  }}>
                    <div>
                      <p style={{ fontWeight: "700", fontSize: "16px", color: "#111827", margin: "0 0 4px" }}>
                        {w.train_name}
                      </p>
                      <p style={{ fontSize: "12px", color: "#2563EB", fontWeight: "600", margin: "0 0 4px" }}>
                        👤 {w.user_name}
                      </p>
                      <p style={{ fontSize: "13px", color: "#6B7280", margin: "0 0 10px" }}>
                        {w.source} → {w.destination} · {w.journey_date}
                      </p>
                      <span style={{
                        background: "#FEF3C7", color: "#92400E",
                        border: "1px solid #FCD34D",
                        borderRadius: "20px", padding: "3px 10px",
                        fontSize: "12px", fontWeight: "700"
                      }}>
                        Position #{w.position}
                      </span>
                    </div>
                    <div style={{ textAlign: "right" }}>
                      <p style={{
                        fontSize: "11px", color: "#9CA3AF",
                        fontWeight: "600", margin: "0 0 4px",
                        textTransform: "uppercase"
                      }}>
                        Confirmation Chance
                      </p>
                      <p style={{
                        fontSize: "28px", fontWeight: "700", margin: 0,
                        color: w.probability >= 0.8 ? "#059669"
                          : w.probability >= 0.5 ? "#D97706" : "#EF4444"
                      }}>
                        {Math.round(w.probability * 100)}%
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default BookingHistory;