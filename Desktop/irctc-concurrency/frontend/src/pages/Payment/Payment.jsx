import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { processPayment } from "../../api/paymentApi";
import Loader from "../../components/Loader";
import { PAYMENT_METHODS, MESSAGES } from "../../constants/constants";
import { formatPrice } from "../../helpers/helpers";

const Payment = () => {
  const location = useLocation();
  const navigate  = useNavigate();


  const { booking, bookings, train, journeyDate, seat, seats } = location.state || {};

  const allBookings = bookings || (booking ? [{ booking, seat }] : []);
  const allSeats    = seats || (seat ? [seat] : []);
  const totalAmount = allBookings.length * (train?.price || 0);

  const [method, setMethod]   = useState("MOCK_UPI");
  const [loading, setLoading] = useState(false);
  const [error, setError]     = useState("");

  const handlePayment = async () => {
    setLoading(true);
    setError("");
    try {
      const paymentResults = [];
      for (const item of allBookings) {
        const b = item.booking || item;
        const res = await processPayment({
          booking_id:     b.booking_id,
          user_id:        b.user_id || 1,
          amount:         train.price,
          payment_method: method,
        });
        if (res.success) {
          paymentResults.push(res.data);
        } else {
          setError(MESSAGES.PAYMENT_FAILED);
          setLoading(false);
          return;
        }
      }

      navigate("/confirmation", {
        state: {
          payments:    paymentResults,
          payment:     paymentResults[0], 
          train,
          journeyDate,
          seats:       allSeats,
          seat:        allSeats[0],
          totalAmount,
        }
      });

    } catch {
      setError(MESSAGES.SERVER_ERROR);
    } finally {
      setLoading(false);
    }
  };

  if (allBookings.length === 0) { navigate("/search"); return null; }

  return (
    <div style={{ maxWidth: "520px", margin: "40px auto", padding: "0 16px" }}>
      <div style={{
        background: "white", borderRadius: "16px",
        padding: "32px", boxShadow: "0 4px 20px rgba(0,0,0,0.1)"
      }}>
        <h1 style={{ fontSize: "22px", fontWeight: "700", marginBottom: "8px" }}>
          💳 Payment
        </h1>
        <p style={{ color: "#6B7280", fontSize: "14px", marginBottom: "28px" }}>
          Complete your booking payment
        </p>

        {/* Booking summary */}
        <div style={{
          background: "#F9FAFB", borderRadius: "10px",
          padding: "16px", marginBottom: "24px"
        }}>
          <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "8px" }}>
            <span style={{ color: "#6B7280", fontSize: "13px" }}>Train</span>
            <span style={{ fontWeight: "600", fontSize: "13px" }}>{train?.train_name}</span>
          </div>

          {/* Show all seats */}
          <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "8px" }}>
            <span style={{ color: "#6B7280", fontSize: "13px" }}>
              {allSeats.length > 1 ? "Seats" : "Seat"}
            </span>
            <span style={{ fontWeight: "600", fontSize: "13px" }}>
              {allSeats.map(s => s.seat_number).join(", ")}
            </span>
          </div>

          <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "8px" }}>
            <span style={{ color: "#6B7280", fontSize: "13px" }}>Date</span>
            <span style={{ fontWeight: "600", fontSize: "13px" }}>{journeyDate}</span>
          </div>

          {/* Per seat price if multiple */}
          {allSeats.length > 1 && (
            <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "8px" }}>
              <span style={{ color: "#6B7280", fontSize: "13px" }}>Price per seat</span>
              <span style={{ fontWeight: "600", fontSize: "13px" }}>{formatPrice(train?.price)}</span>
            </div>
          )}

          <div style={{
            display: "flex", justifyContent: "space-between",
            paddingTop: "12px", borderTop: "1px solid #E5E7EB", marginTop: "8px"
          }}>
            <span style={{ fontWeight: "700" }}>
              Total {allSeats.length > 1 ? `(${allSeats.length} seats)` : ""}
            </span>
            <span style={{ fontWeight: "700", fontSize: "18px", color: "#1E3A5F" }}>
              {formatPrice(totalAmount)}
            </span>
          </div>
        </div>

        {/* Payment method */}
        <div style={{ marginBottom: "24px" }}>
          <p style={{ fontWeight: "600", fontSize: "13px", marginBottom: "12px", color: "#374151" }}>
            SELECT PAYMENT METHOD
          </p>
          <div style={{ display: "flex", flexDirection: "column", gap: "10px" }}>
            {PAYMENT_METHODS.map((m) => (
              <label key={m.value} style={{
                display: "flex", alignItems: "center", gap: "12px",
                padding: "12px 16px", borderRadius: "8px", cursor: "pointer",
                border: method === m.value ? "2px solid #2563EB" : "2px solid #E5E7EB",
                background: method === m.value ? "#EFF6FF" : "white",
              }}>
                <input
                  type="radio" name="payment"
                  value={m.value} checked={method === m.value}
                  onChange={() => setMethod(m.value)}
                  style={{ accentColor: "#2563EB" }}
                />
                <span style={{ fontWeight: "500", fontSize: "14px" }}>{m.label}</span>
              </label>
            ))}
          </div>
        </div>

        {/* Error */}
        {error && (
          <div style={{
            background: "#FEF2F2", border: "1px solid #FECACA",
            borderRadius: "8px", padding: "12px", color: "#DC2626",
            fontSize: "13px", marginBottom: "16px"
          }}>
             {error}
          </div>
        )}

        {/* Pay button */}
        {loading ? <Loader message="Processing payment..." /> : (
          <button
            onClick={handlePayment}
            style={{
              width: "100%", background: "#059669",
              color: "white", border: "none",
              borderRadius: "10px", padding: "14px",
              fontWeight: "700", fontSize: "16px",
              cursor: "pointer",
            }}
          >
            Pay {formatPrice(totalAmount)}
          </button>
        )}
      </div>
    </div>
  );
};

export default Payment;