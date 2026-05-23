import { useLocation, useNavigate } from "react-router-dom";
import { formatPrice } from "../../helpers/helpers";

const Confirmation = () => {
  const location = useLocation();
  const navigate  = useNavigate();
  const { payment, payments, train, journeyDate, seat, seats, totalAmount } = location.state || {};

  const allPayments = payments || (payment ? [payment] : []);
  const allSeats    = seats || (seat ? [seat] : []);

  if (allPayments.length === 0) { navigate("/search"); return null; }

  const total = totalAmount || (train?.price * allPayments.length);

  return (
    <div style={{ maxWidth: "520px", margin: "40px auto", padding: "0 16px" }}>
      <div style={{
        background: "white", borderRadius: "16px",
        padding: "32px", boxShadow: "0 4px 20px rgba(0,0,0,0.1)",
        textAlign: "center"
      }}>
        {/* Success icon */}
        <div style={{
          width: "72px", height: "72px",
          background: "#D1FAE5", borderRadius: "50%",
          display: "flex", alignItems: "center", justifyContent: "center",
          margin: "0 auto 20px", fontSize: "32px"
        }}>
          
        </div>

        <h1 style={{ fontSize: "22px", fontWeight: "700", marginBottom: "8px", color: "#059669" }}>
          Booking Confirmed!
        </h1>
        <p style={{ color: "#6B7280", marginBottom: "28px" }}>
          {allPayments.length > 1
            ? `${allPayments.length} tickets booked successfully`
            : "Your ticket has been booked successfully"}
        </p>

        {/* PNR box — show all PNRs */}
        {allPayments.map((p, i) => (
          <div key={i} style={{
            background: "#1E3A5F", borderRadius: "12px",
            padding: "16px 20px", marginBottom: "12px", color: "white"
          }}>
            {allPayments.length > 1 && (
              <p style={{ opacity: 0.7, fontSize: "11px", margin: "0 0 4px" }}>
                TICKET {i + 1} — Seat {allSeats[i]?.seat_number}
              </p>
            )}
            <p style={{ opacity: 0.7, fontSize: "12px", margin: "0 0 6px", letterSpacing: "0.1em" }}>
              PNR NUMBER
            </p>
            <p style={{ fontSize: "24px", fontWeight: "700", margin: 0, letterSpacing: "0.05em" }}>
              {p.pnrNumber}
            </p>
          </div>
        ))}

        {/* Details */}
        <div style={{
          background: "#F9FAFB", borderRadius: "10px",
          padding: "16px", marginBottom: "24px", textAlign: "left"
        }}>
          {[
            ["Train",        train?.train_name],
            ["Seats",        allSeats.map(s => s.seat_number).join(", ")],
            ["Journey Date", journeyDate],
            ["Amount Paid",  formatPrice(total)],
            ["Payment ID",   allPayments[0]?.transactionId],
          ].map(([label, value]) => (
            <div key={label} style={{
              display: "flex", justifyContent: "space-between",
              padding: "8px 0", borderBottom: "1px solid #E5E7EB"
            }}>
              <span style={{ color: "#6B7280", fontSize: "13px" }}>{label}</span>
              <span style={{ fontWeight: "600", fontSize: "13px" }}>{value}</span>
            </div>
          ))}
        </div>

        <button
          onClick={() => navigate("/search")}
          style={{
            width: "100%", background: "#2563EB",
            color: "white", border: "none",
            borderRadius: "10px", padding: "14px",
            fontWeight: "700", fontSize: "15px", cursor: "pointer"
          }}
        >
          Book Another Ticket
        </button>
      </div>
    </div>
  );
};

export default Confirmation;